// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package service

import (
	"context"
	"encoding/base64"

	csapi "github.com/gitpod-io/gitpod/content-service/api"
	"github.com/gitpod-io/gitpod/content-service/pkg/layer"
	regapi "github.com/gitpod-io/gitpod/registry-facade/api"
	workspacev1 "github.com/gitpod-io/gitpod/ws-manager/api/crd/v1"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkspaceImageSpecProvider struct {
	Client          client.Client
	Namespace       string
	ContentProvider *layer.Provider

	regapi.UnimplementedSpecProviderServer
}

func (is *WorkspaceImageSpecProvider) GetImageSpec(ctx context.Context, req *regapi.GetImageSpecRequest) (*regapi.GetImageSpecResponse, error) {
	var ws workspacev1.Workspace
	err := is.Client.Get(ctx, types.NamespacedName{Namespace: is.Namespace, Name: req.Id}, &ws)
	if errors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	spec := &regapi.ImageSpec{
		BaseRef:       pointer.StringDeref(ws.Spec.Image.Workspace.Ref, ""),
		IdeRef:        ws.Spec.Image.IDE.Web,
		DesktopIdeRef: ws.Spec.Image.IDE.Desktop,
		SupervisorRef: ws.Spec.Image.IDE.Supervisor,
	}

	initializerPB, err := base64.StdEncoding.DecodeString(string(ws.Spec.Initializer))
	if err != nil {
		return nil, xerrors.Errorf("cannot decode init config: %w", err)
	}
	var initializer csapi.WorkspaceInitializer
	err = proto.Unmarshal(initializerPB, &initializer)
	if err != nil {
		return nil, xerrors.Errorf("cannot unmarshal init config: %w", err)
	}
	cl, _, err := is.ContentProvider.GetContentLayerPVC(ctx, ws.Spec.Ownership.Owner, ws.Name, &initializer)
	if err != nil {
		return nil, xerrors.Errorf("cannot get content layer: %w", err)
	}

	contentLayer := make([]*regapi.ContentLayer, len(cl))
	for i, l := range cl {
		if len(l.Content) > 0 {
			contentLayer[i] = &regapi.ContentLayer{
				Spec: &regapi.ContentLayer_Direct{
					Direct: &regapi.DirectContentLayer{
						Content: l.Content,
					},
				},
			}
			continue
		}

		diffID := l.DiffID
		contentLayer[i] = &regapi.ContentLayer{
			Spec: &regapi.ContentLayer_Remote{
				Remote: &regapi.RemoteContentLayer{
					DiffId:    diffID,
					Digest:    l.Digest,
					MediaType: string(l.MediaType),
					Url:       l.URL,
					Size:      l.Size,
				},
			},
		}
	}
	spec.ContentLayer = contentLayer

	return &regapi.GetImageSpecResponse{
		Spec: spec,
	}, nil
}
