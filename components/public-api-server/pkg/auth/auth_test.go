// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/mattbarton/go-cookie-signature/signature"
	"github.com/stretchr/testify/require"
)

func TestBearerTokenFromHeaders(t *testing.T) {
	type Scenario struct {
		Name string

		// Input
		Header http.Header

		// Output
		Token string
		Error error
	}

	for _, s := range []Scenario{
		{
			Name:   "happy case",
			Header: addToHeader(http.Header{}, "Authorization", "Bearer foo"),
			Token:  "foo",
		},
		{
			Name:   "leading and trailing spaces are trimmed",
			Header: addToHeader(http.Header{}, "Authorization", "  Bearer foo  "),
			Token:  "foo",
		},
		{
			Name:   "anything after Bearer is extracted",
			Header: addToHeader(http.Header{}, "Authorization", "Bearer foo bar"),
			Token:  "foo bar",
		},
		{
			Name:   "duplicate bearer",
			Header: addToHeader(http.Header{}, "Authorization", "Bearer Bearer foo"),
			Token:  "Bearer foo",
		},
		{
			Name:   "missing Bearer prefix fails",
			Header: addToHeader(http.Header{}, "Authorization", "foo"),
			Error:  NoAccessToken,
		},
		{
			Name:   "missing Authorization header fails",
			Header: http.Header{},
			Error:  NoAccessToken,
		},
	} {
		t.Run(s.Name, func(t *testing.T) {
			token, err := BearerTokenFromHeaders(s.Header)
			require.ErrorIs(t, err, s.Error)
			require.Equal(t, s.Token, token)
		})
	}
}

func addToHeader(h http.Header, key, value string) http.Header {
	h.Add(key, value)
	return h
}

func TestTokenFromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://some-endpoint.com", nil)
	session := uuid.New()
	signingSecret := "Important!Really-Change-This-Key!"
	cookie := &http.Cookie{
		Name:     "_gitpod_io_",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Value:    url.QueryEscape(fmt.Sprintf("s:%s", signature.Sign(session.String(), signingSecret))),
	}
	req.AddCookie(cookie)

	token, err := TokenFromCookie(cookie.Name, signingSecret, req.Header)

	require.NoError(t, err)
	require.Equal(t, session.String(), token)

}
