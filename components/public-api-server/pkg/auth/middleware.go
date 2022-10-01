// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package auth

import (
	"context"

	"github.com/bufbuild/connect-go"
)

func NewInterceptor(cookieName, cookieSigningSecret string) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {

			if req.Spec().IsClient {
				return next(ctx, req)
			}

			headers := req.Header()

			token, err := BearerTokenFromHeaders(headers)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)

			}
			return next(TokenToContext(ctx, token), req)

			// token, err := TokenFromCookie(cookieName, cookieSigningSecret, headers)
			// if err == nil {
			// 	return next(TokenToContext(ctx, token), req)
			// }

		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
