package token

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestExampleFromDocs is shown at
// https://pkg.go.dev/net/http/httptest#example-ResponseRecorder

var token *Token
var err error

func init() {
	token, err = NewToken(
		"https://exampletest.com",
		"XXXXXclientidXXXXX",
		"XXXXXclientsecretXXXXX",
		[]string{"offline_access", "accounting.transactions"},
		"", // authURL
		"", // tokenURL
	)
	if err != nil {
		log.Fatalf("token initialisation failed")
	}
}

func TestExampleFromDocs(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	// body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("Status code %d != 200", resp.StatusCode)
	}
}

// Test home page
func TestHandleHome(t *testing.T) {

	handler := token.HandleHome

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	bodyString := string(body)

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}
	if !strings.Contains(bodyString, "<h4>Code generation</h4>") {
		t.Errorf("body content unexpected")
	}
}

// Test home page redirecting to code with an incorrect state
func TestHandleHomeRedirectCodeErrorState(t *testing.T) {

	handler := token.HandleHome

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/?code=abc", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	statusCode := resp.StatusCode

	if statusCode != 403 {
		t.Errorf("Status code %d != 403", statusCode)
	}
}

func TestHandleHomeRedirectCode(t *testing.T) {

	handler := token.HandleHome

	fragment := fmt.Sprintf("?code=%s&state=%s", "123", token.state)
	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/"+fragment, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	statusCode := resp.StatusCode

	// redirect to code
	if statusCode != 302 {
		t.Errorf("Status code %d != 302", statusCode)
	}
}

func TestHandleCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	handler := token.HandleCode

	fragment := fmt.Sprintf("?code=%s", "123")
	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/"+fragment, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	bodyString := string(body)

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}
	if !strings.Contains(bodyString, "Code extraction succeeded") {
		t.Errorf("body content unexpected")
	}
	if token.AccessToken != "abc" {
		t.Errorf("access token value unexpected: %s", token.AccessToken)
	}
	if token.RefreshToken != "def" {
		t.Errorf("refresh token value unexpected: %s", token.RefreshToken)
	}
}

func TestHandleCodeFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	handler := token.HandleCode

	fragment := fmt.Sprintf("?code=%s", "123")
	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/"+fragment, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	statusCode := resp.StatusCode

	if statusCode != 503 {
		t.Errorf("Status code %d != 503", statusCode)
	}
}

func TestHandleRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "hij", "refresh_token": "klm", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	handler := token.HandleRefresh

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/refresh", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	statusCode := resp.StatusCode

	// redirect to token
	if statusCode != 302 {
		t.Errorf("Status code %d != 302", statusCode)
	}
}

func TestHandleRefreshFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	handler := token.HandleRefresh

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/refresh", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	statusCode := resp.StatusCode

	if statusCode != 503 {
		t.Errorf("Status code %d != 503", statusCode)
	}
}

func TestHandleToken(t *testing.T) {

	token.AccessToken = "xyz123"
	handler := token.HandleAccessToken

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/token", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "application/json" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}

	var r map[string]string
	json.Unmarshal(body, &r)
	at, ok := r["accessToken"]
	if !ok {
		t.Error("No accessToken in results")
	}
	if at != token.AccessToken {
		t.Errorf("AccessToken is %s should be %s", at, token.AccessToken)
	}
}

func TestHandleRefreshToken(t *testing.T) {

	token.RefreshToken = "abc987"
	handler := token.HandleRefreshToken

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/refresh", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "application/json" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}

	var r map[string]string
	json.Unmarshal(body, &r)
	at, ok := r["refreshToken"]
	if !ok {
		t.Error("No refreshToken in results")
	}
	if at != token.RefreshToken {
		t.Errorf("RefreshToken is %s should be %s", at, token.RefreshToken)
	}
}

func TestHandleRefreshTokenFail(t *testing.T) {

	token.RefreshToken = ""
	handler := token.HandleRefreshToken

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/refresh", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	statusCode := resp.StatusCode

	if statusCode != 405 {
		t.Errorf("Status code %d != 200", statusCode)
	}
}
