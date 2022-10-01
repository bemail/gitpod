// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mattbarton/go-cookie-signature/signature"
)

var (
	NoAccessToken      = errors.New("missing access token")
	InvalidAccessToken = errors.New("invalid access token")
)

func BearerTokenFromHeaders(h http.Header) (string, error) {
	const bearerPrefix = "Bearer "

	authorization := strings.TrimSpace(h.Get("Authorization"))
	if authorization == "" {
		return "", fmt.Errorf("empty authorization header: %w", NoAccessToken)
	}

	if !strings.HasPrefix(authorization, bearerPrefix) {
		return "", fmt.Errorf("authorization header does not have a Bearer prefix: %w", NoAccessToken)
	}

	return strings.TrimPrefix(authorization, bearerPrefix), nil
}

func TokenFromCookie(cookieName string, signingSecret string, h http.Header) (string, error) {
	// We need to parse cookies, but we don't have access to an http.Request to call Cookies(), so we create a fake request.
	req := http.Request{Header: h}
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("token not present in cookie: %w", NoAccessToken)
	}

	// By default, cookies set by the `server` component are URL encoded
	// https://github.com/jshttp/cookie#encode
	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("cannot query unescape token: %w", InvalidAccessToken)
	}

	// express prefixes the cookie value with `s:`, we need to remove that.
	// https://github.com/expressjs/session/blob/1010fadc2f071ddf2add94235d72224cf65159c6/index.js#L656
	value = strings.TrimPrefix(value, "s:")

	token, valid := signature.Unsign(value, signingSecret)
	if !valid {
		return "", fmt.Errorf("cookie does not have a valid signature: %w", InvalidAccessToken)
	}

	return token, nil
}
