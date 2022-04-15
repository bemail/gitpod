// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package supervisor

import (
	"context"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/gitpod-io/gitpod/supervisor/api"
)

const (
	SubscriptionTypeNotify = iota
	SubscriptionTypeAction = iota
)

// NewNotificationService creates a new notification service.
func NewNotificationService() *NotificationService {
	return &NotificationService{
		helper: NewRPCStreamHelper(),
	}
}

// NotificationService implements the notification service API.
type NotificationService struct {
	m                    sync.Mutex
	helper               *RPCStreamHelper
	latestActiveClientID uint64

	api.UnimplementedNotificationServiceServer
}

// RegisterGRPC registers a gRPC service.
func (srv *NotificationService) RegisterGRPC(s *grpc.Server) {
	api.RegisterNotificationServiceServer(s, srv)
}

// RegisterREST registers a REST service.
func (srv *NotificationService) RegisterREST(mux *runtime.ServeMux, grpcEndpoint string) error {
	return api.RegisterNotificationServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, []grpc.DialOption{grpc.WithInsecure()})
}

// Notify sends a notification to the user.
func (srv *NotificationService) Notify(ctx context.Context, req *api.NotifyRequest) (*api.NotifyResponse, error) {
	d, err := srv.helper.DoAction(ctx, SubscriptionTypeNotify, len(req.Actions) != 0, func(requestID uint64) interface{} {
		return &api.SubscribeResponse{
			RequestId: requestID,
			Request:   req,
		}
	})
	if err != nil {
		return nil, err
	}
	t, ok := d.(*api.NotifyResponse)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Parse response failed")
	}
	return t, err
}

// Subscribe subscribes to notifications that are sent to the supervisor.
func (srv *NotificationService) Subscribe(req *api.SubscribeRequest, resp api.NotificationService_SubscribeServer) error {
	log.WithField("SubscribeRequest", req).Info("Subscribe entered")
	defer log.WithField("SubscribeRequest", req).Info("Subscribe exited")
	subscriber := srv.helper.Subscribe(resp.Context(), SubscriptionTypeNotify)
	return srv.helper.Wait(resp.Context(), subscriber, func(d interface{}) error {
		data, ok := d.(*api.SubscribeResponse)
		if !ok {
			return status.Errorf(codes.Internal, "Parse response failed")
		}
		return resp.Send(data)
	})
}

// Respond reports user actions as response to a notification request.
func (srv *NotificationService) Respond(ctx context.Context, req *api.RespondRequest) (*api.RespondResponse, error) {
	err := srv.helper.Respond(ctx, req.RequestId, req.Response, func(msg interface{}) error {
		t, ok := msg.(*api.SubscribeResponse)
		if !ok {
			return status.Errorf(codes.Internal, "Parse response failed")
		}
		if !isActionAllowed(req.Response.Action, t.Request) {
			log.WithFields(map[string]interface{}{
				"Notification": t,
				"Action":       req.Response.Action,
			}).Error("Invalid user action on notification")
			return status.Errorf(codes.InvalidArgument, "Invalid user action on notification")
		}
		return nil
	})
	return &api.RespondResponse{}, err
}

func isActionAllowed(action string, req *api.NotifyRequest) bool {
	if action == "" {
		// user cancelled, which is always allowed
		return true
	}
	for _, allowedAction := range req.Actions {
		if allowedAction == action {
			return true
		}
	}
	return false
}

func (srv *NotificationService) SetActiveClient(req *api.SetActiveClientRequest, resp api.NotificationService_SetActiveClientServer) error {
	log.WithField("SetActiveClientRequest", req).Info("SetActiveClient entered")
	defer log.WithField("SetActiveClientRequest", req).Info("SetActiveClient exited")
	subscriber := srv.helper.Subscribe(resp.Context(), SubscriptionTypeAction)
	srv.m.Lock()
	if srv.latestActiveClientID != 0 && subscriber.id != srv.latestActiveClientID {
		srv.helper.Unsubscribe(srv.latestActiveClientID)
		srv.latestActiveClientID = subscriber.id
	}
	srv.m.Unlock()
	return srv.helper.Wait(resp.Context(), subscriber, func(d interface{}) error {
		t, ok := d.(*api.SetActiveClientResponse)
		if !ok {
			return status.Errorf(codes.Internal, "Parse response failed")
		}
		return resp.Send(t)
	})
}

func (srv *NotificationService) Action(ctx context.Context, req *api.ActionRequest) (*api.ActionResponse, error) {
	wait := true
	d, err := srv.helper.DoAction(ctx, SubscriptionTypeAction, wait, func(requestID uint64) interface{} {
		return &api.SetActiveClientResponse{
			RequestId: requestID,
			Request:   req,
		}
	})
	if err != nil {
		return nil, err
	}
	t, ok := d.(*api.ActionResponse)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Parse response failed")
	}
	return t, err
}

func (srv *NotificationService) ActionRespond(ctx context.Context, req *api.ActionRespondRequest) (*api.ActionRespondResponse, error) {
	return &api.ActionRespondResponse{}, srv.helper.Respond(ctx, req.RequestId, req.Response, func(msg interface{}) error {
		return nil
	})
}
