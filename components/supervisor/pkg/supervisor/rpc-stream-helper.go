// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package supervisor

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/gitpod-io/gitpod/supervisor/api"
)

const (
	MaxPending = 120
	MaxSuber   = 100
)

// NewRPCStreamHelper creates a new RPCStreamHelper service.
func NewRPCStreamHelper() *RPCStreamHelper {
	return &RPCStreamHelper{
		subscriptions:   make(map[uint64]*subscription),
		pendingRequests: make(map[uint64]*pendingRequest),
	}
}

// RPCStreamHelper to Subscribe gRPC response streams to ask them to DoAction and wait them to Respond
type RPCStreamHelper struct {
	mutex              sync.Mutex
	subscriptions      map[uint64]*subscription
	pendingRequests    map[uint64]*pendingRequest
	nextRequestID      uint64
	nextSubscriptionID uint64
}

type pendingRequest struct {
	requestID       uint64
	wait            bool
	message         interface{}
	responseChannel chan interface{}
	once            sync.Once
	closed          bool
}

func (pending *pendingRequest) close() {
	pending.once.Do(func() {
		close(pending.responseChannel)
		pending.closed = true
	})
}

type subscription struct {
	id      uint64
	subType int
	channel chan interface{}
	once    sync.Once
	closed  bool
	cancel  context.CancelFunc
}

func (subscription *subscription) close() {
	subscription.once.Do(func() {
		close(subscription.channel)
		subscription.closed = true
		subscription.cancel()
	})
}

type messageHandler func(requestID uint64) interface{}

// DoAction ask client to DoAction.
// Will Hold all requests if no subscription is registered until got at least one subscription.
// If the wait is false, will respond with empty data immediately
func (srv *RPCStreamHelper) DoAction(ctx context.Context, subType int, wait bool, messageHandler messageHandler) (interface{}, error) {
	if len(srv.pendingRequests) >= MaxPending {
		return nil, status.Error(codes.ResourceExhausted, "Max number of pending DoAction request exceeded")
	}

	srv.mutex.Lock()
	var (
		requestID = srv.nextRequestID
	)
	srv.nextRequestID++
	srv.mutex.Unlock()

	message := messageHandler(requestID)
	pending := srv.notifySubscribers(requestID, subType, wait, message)
	select {
	case resp, ok := <-pending.responseChannel:
		if !ok {
			log.Error("Do action channel has been closed")
			return nil, status.Error(codes.Aborted, "response channel closed")
		}
		log.WithField("Response", resp).Info("DoAction response")
		return resp, nil
	case <-ctx.Done():
		log.Info("DoAction cancelled")
		srv.mutex.Lock()
		defer srv.mutex.Unlock()
		// make sure the DoAction request has not been responded in between these selectors
		_, ok := srv.pendingRequests[pending.requestID]
		if ok {
			delete(srv.pendingRequests, pending.requestID)
			pending.close()
		}
		return nil, ctx.Err()
	}
}

func (srv *RPCStreamHelper) notifySubscribers(requestID uint64, subType int, wait bool, message interface{}) *pendingRequest {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	srv.nextRequestID++
	for _, subscription := range srv.subscriptions {
		if subscription.subType != subType {
			continue
		}
		select {
		case subscription.channel <- message:
			// all good
		default:
			// subscriber doesn't consume messages fast enough
			log.Info("Cancelling unresponsive subscriber")
			delete(srv.subscriptions, subscription.id)
			subscription.close()
		}
	}
	channel := make(chan interface{}, 1)
	pending := &pendingRequest{
		message:         message,
		responseChannel: channel,
		requestID:       requestID,
		wait:            wait,
	}
	srv.pendingRequests[requestID] = pending
	if !wait {
		// produce an immediate response
		channel <- &api.NotifyResponse{}
		pending.close()
	}
	return pending
}

type sendRespHandler func(d interface{}) error

// Subscribe a subscription.
func (srv *RPCStreamHelper) Subscribe(respCtx context.Context, subType int) *subscription {
	subscription := srv.subscribeLocked(respCtx, subType)
	return subscription
}

// Wait to hold response stream to ask client to DoAction
func (srv *RPCStreamHelper) Wait(respCtx context.Context, subscription *subscription, sendRespHandler sendRespHandler) error {
	defer srv.Unsubscribe(subscription.id)
	for {
		select {
		case subscribeResponse, ok := <-subscription.channel:
			if !ok || subscription.closed {
				return status.Errorf(codes.Aborted, "Subscriber channel closed.")
			}
			err := sendRespHandler(subscribeResponse)
			if err != nil {
				return status.Errorf(codes.Internal, "Sending DoAction request failed. %s", err)
			}
		case <-respCtx.Done():
			log.Info("Subscriber cancelled")
			return nil
		}
	}
}

// Unsubscribe a subscription.
func (srv *RPCStreamHelper) Unsubscribe(subscriberID uint64) {
	srv.unsubscribeLocked(subscriberID)
}

func (srv *RPCStreamHelper) subscribeLocked(ctx context.Context, subType int) *subscription {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	// account for some back pressure
	capacity := len(srv.pendingRequests)
	if MaxSuber > capacity {
		capacity = MaxSuber
	}
	channel := make(chan interface{}, capacity)
	log.WithField("pending", len(srv.pendingRequests)).Info("Sending pending DoAction requests")
	for id, pending := range srv.pendingRequests {
		channel <- pending.message
		if !pending.wait {
			delete(srv.pendingRequests, id)
		}
	}
	id := srv.nextSubscriptionID
	srv.nextSubscriptionID++
	_, cancel := context.WithCancel(ctx)
	subscription := &subscription{
		subType: subType,
		channel: channel,
		id:      id,
		cancel:  cancel,
	}
	srv.subscriptions[id] = subscription
	return subscription
}

func (srv *RPCStreamHelper) unsubscribeLocked(subscriptionID uint64) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	subscription, ok := srv.subscriptions[subscriptionID]
	if !ok {
		log.Errorf("Could not unsubscribe subscriber")
		return
	}
	delete(srv.subscriptions, subscription.id)
	subscription.close()
}

type checkRespHandler func(msg interface{}) error

// Respond result as request data to respond the client that ask to DoAction.
func (srv *RPCStreamHelper) Respond(ctx context.Context, requestID uint64, reqResp interface{}, checkRespHandler checkRespHandler) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	pending, ok := srv.pendingRequests[requestID]
	if !ok {
		log.Info("Invalid or late response to DoAction request")
		return status.Errorf(codes.DeadlineExceeded, "Invalid or late response to DoAction request")
	}
	if err := checkRespHandler(pending.message); err != nil {
		log.WithError(err).Error("Invalid user action on DoAction request")
		return status.Errorf(codes.InvalidArgument, "Invalid user action on DoAction request")
	}
	if !pending.closed {
		pending.responseChannel <- reqResp
		pending.close()
	}
	delete(srv.pendingRequests, requestID)
	return nil
}
