package token

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// depends on token and err in handler_test

func TestNewTokenErr(t *testing.T) {

	type newTokenInput struct {
		redirect string
		client   string
		secret   string
		scopes   []string
		authURL  string
		tokenURL string
	}

	// https://www.myhatchpad.com/insight/mocking-techniques-for-go/
	tests := []struct {
		name        string
		input       *newTokenInput
		expectedErr error
	}{
		{
			name: "empty_redirect",
			input: &newTokenInput{
				redirect: "",
				client:   "abc",
				secret:   "def",
				scopes:   []string{},
				authURL:  "",
				tokenURL: "",
			},
			expectedErr: errors.New("redirect url invalid"),
		},
		{
			name: "empty_client",
			input: &newTokenInput{
				redirect: "http://xero.com",
				client:   "",
				secret:   "def",
				scopes:   []string{},
				authURL:  "",
				tokenURL: "",
			},
			expectedErr: errors.New("redirect, client or secret is empty"),
		},
		{
			name: "empty_secret",
			input: &newTokenInput{
				redirect: "http://xero.com/",
				client:   "abc",
				secret:   "",
				scopes:   []string{},
				authURL:  "",
				tokenURL: "",
			},
			expectedErr: errors.New("redirect, client or secret is empty"),
		},
		{
			name: "empty_scopes",
			input: &newTokenInput{
				redirect: "http://xero.com/",
				client:   "abc",
				secret:   "def",
				scopes:   []string{},
				authURL:  "",
				tokenURL: "",
			},
			expectedErr: errors.New("scopes cannot be empty"),
		},
		{
			name: "ok_scopes",
			input: &newTokenInput{
				redirect: "http://xero.com/",
				client:   "abc",
				secret:   "def",
				scopes:   []string{"offline_access", "accounting.transactions"},
				authURL:  "",
				tokenURL: "",
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewToken(
				test.input.redirect,
				test.input.client,
				test.input.secret,
				test.input.scopes,
				test.input.authURL,
				test.input.tokenURL,
			)
			// nil error match
			if test.expectedErr == nil {
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("expected (%v), got (%v)", test.expectedErr, err)
				}
				// string match
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("expected (%v), got (%v)", test.expectedErr, err)
			}
		})
	}
}

func TestURL(t *testing.T) {

	token.authURL = "http://127.0.0.1:5000/"
	urlForTest := token.AuthURL()
	u, err := url.Parse(urlForTest)
	if err != nil {
		t.Errorf("error parsing url from AuthURL: %s", err)
	}

	args := []string{"response_type", "client_id", "redirect_uri", "scope", "state"}
	params := u.Query()
	scope := ""
	for _, s := range token.Scopes {
		scope += fmt.Sprintf(" %s", s)
	}
	scope = strings.TrimSpace(scope)

	for _, a := range args {
		switch a {
		case "response_type":
			if params[a][0] != "code" {
				t.Errorf("incorrect %s", params[a])
			}
		case "client_id":
			if params[a][0] != token.clientID {
				t.Errorf("incorrect %s", params[a])
			}
		case "redirect_uri":
			if params[a][0] != token.redirectURL {
				t.Errorf("incorrect %s", params[a])
			}
		case "scope":
			if params[a][0] != scope {
				t.Errorf("incorrect have(%s) want(%s)", params[a], scope)
			}
		case "state":
			if params[a][0] != token.state {
				t.Errorf("incorrect %s", params[a])
			}
		}
	}
}

func TestGetToken(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	token.AccessToken = ""
	token.RefreshToken = ""
	err := token.GetToken(token.authURL)

	if err != nil {
		t.Errorf("error %s", err)
	}
	if token.AccessToken == "" {
		t.Errorf("access token is empty")
	}
	if token.RefreshToken == "" {
		t.Errorf("refresh token is empty")
	}

}

func TestGetTokenFail(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	token.AccessToken = ""
	token.RefreshToken = ""
	err := token.GetToken(token.authURL)

	if err.Error() != "empty response received from server" {
		t.Errorf("unexpected error %s", err)
	}
}

func TestGetTokenTimeout(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{"access_token": "ok", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	token.AccessToken = ""
	token.RefreshToken = ""
	token.httpclientTimeout = time.Millisecond * 150
	err := token.GetToken(token.authURL)

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("unexpected error %s", err)
	}
}

func TestRefresh(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	err := token.Refresh()

	if err != nil {
		t.Errorf("error %s", err)
	}
	if token.AccessToken == "" {
		t.Errorf("access token is empty")
	}
	if token.RefreshToken == "" {
		t.Errorf("refresh token is empty")
	}

}

func TestRefreshFail(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	err := token.Refresh()

	if err.Error() != "empty response received from server" {
		t.Errorf("unexpected error %s", err)
	}

}

func TestRefreshFailNonInit(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.RefreshToken = ""
	token.tokenURL = server.URL
	err := token.Refresh()

	if err.Error() != "token system has not been initialised" {
		t.Errorf("unexpected error %s", err)
	}

}
