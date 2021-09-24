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
	"time"
)

// TestExampleFromDocs is shown at
// https://pkg.go.dev/net/http/httptest#example-ResponseRecorder

func initToken() *Token {
	token, err := NewToken(
		"https://exampletest.com",
		"XXXXXclientidXXXXX",
		"XXXXXclientsecretXXXXX",
		[]string{"offline_access", "accounting.transactions"},
		"", // authURL
		"", // tokenURL
		"", // tenantsURL
		10, // refresh minutes
	)
	if err != nil {
		log.Fatalf("token initialisation failed")
	}
	return token
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
	token := initToken()

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
	if strings.Contains(bodyString, "The server is already initialised") {
		t.Errorf("the server should not report being initialised")
	}
	if !strings.Contains(bodyString, "<h4>Code generation</h4>") {
		t.Errorf("body content unexpected")
	}
}

func TestHandleHomeAlreadyInited(t *testing.T) {
	token := initToken()

	handler := token.HandleHome
	token.AccessToken = "abc"
	token.RefreshToken = "def"

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
	if !strings.Contains(bodyString, "The server is already initialised") {
		t.Errorf("the server should report being initialised")
	}
}

// Test home page redirecting to code with an incorrect state
func TestHandleHomeRedirectCodeErrorState(t *testing.T) {
	token := initToken()
	token.state = "123"

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
	token := initToken()

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
	token := initToken()
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
	token := initToken()
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

	if statusCode != 404 {
		t.Errorf("Status code %d != 404", statusCode)
	}
}

func TestHandleRefresh(t *testing.T) {
	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"

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
	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"

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

	if statusCode != 404 {
		t.Errorf("Status code %d != 404", statusCode)
	}
}

func TestHandleToken(t *testing.T) {
	token := initToken()
	token.AccessToken = "xyz123"
	token.AccessTokenExpiryUTC = time.Now().UTC().Add(time.Minute * 10)
	token.RefreshToken = "abc987"
	token.RefreshTokenExpiryUTC = time.Now().UTC().Add(time.Hour * 10)

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
		t.Errorf("body: %s", body)
		t.Errorf("token: %s", token)
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

func TestHandleTokenFailOld(t *testing.T) {
	token := initToken()
	token.AccessToken = "xyz123"
	token.RefreshToken = "abc987"
	// expiration times are at the go epoch

	handler := token.HandleAccessToken

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/token", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	statusCode := resp.StatusCode

	if statusCode != 500 {
		t.Errorf("Status code %d != 500", statusCode)
		t.Errorf("body: %s", body)
	}
}

func TestHandleRefreshToken(t *testing.T) {
	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"

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
	token := initToken()

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

func TestHandleStatus(t *testing.T) {
	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	token.AccessTokenExpiryUTC = time.Now().UTC().Add(time.Minute * 30)
	token.RefreshTokenExpiryUTC = time.Now().UTC().Add(time.Hour * 24 * 30)

	handler := token.HandleStatus

	req := httptest.NewRequest("GET", "http://127.0.0.1:5001/status", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(resp.Body)

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "application/json" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}

	var r map[string]string
	json.Unmarshal(body, &r)
	at, ok := r["refresh_token"]
	if !ok {
		t.Error("No refreshToken in results")
	}
	if at != token.RefreshToken {
		t.Errorf("RefreshToken is %s should be %s", at, token.RefreshToken)
	}
	at, ok = r["access_token"]
	if !ok {
		t.Error("No access token in results")
	}
	if at != token.AccessToken {
		t.Errorf("AccessToken is %s should be %s", at, token.AccessToken)
	}
}

func TestHandleTenantsFail(t *testing.T) {

	failResponse := `{"Type":null,"Title":"Unauthorized","Status":401,"Detail":"TokenExpired: token expired at 09/24/2021 08:23:47","Instance":"04115443-c574-4949-9cfa-102bdb03ca91","Extensions":{}}`

	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(failResponse))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	handler := token.HandleTenants

	// req := httptest.NewRequest("GET", "http://127.0.0.1/tenants", nil)
	req := httptest.NewRequest("GET", server.URL, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(resp.Body)

	if statusCode != 500 {
		t.Errorf("Status code %d != 500", statusCode)
	}
	// if contentType != "application/json" {
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}
	if !strings.Contains(string(body), "Tenant callout http error, 401") {
		t.Error("Body content unexpected for 403 status message")
	}

}

func TestHandleTenantsSucceed(t *testing.T) {

	okResponse := `[
		{
			"ID":"5010b97c-1d4f-11ec-821e-4756fdf46484",
			 "AuthEventID":"55c8fb68-1d4f-11ec-aeb7-3b421a2838d0",
			 "TenantID":"87d9664c-1d4f-11ec-ba64-af1d5e634f86",
			 "TenantType":"ORGANISATION",
			 "TenantName":"My Glorious Organization",
			 "CreatedDateUTC":"2021-07-01T20:55:00.5717400",
			 "UpdatedDateUTC":"2021-09-21T19:44:35.5610020"
		 }
	]`

	token := initToken()
	token.AccessToken = "abc"
	token.RefreshToken = "def"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(okResponse))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	handler := token.HandleTenants

	req := httptest.NewRequest("GET", server.URL, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(resp.Body)

	if statusCode != 200 {
		t.Errorf("Status code %d != 200", statusCode)
	}
	if contentType != "application/json" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}

	var tenants *Tenants
	err := json.Unmarshal(body, &tenants)
	if err != nil {
		t.Errorf("unexpected tenants error: %s", err)
	}
	if len(*tenants) != 1 {
		t.Errorf("unexpected tenants length want(1), got (%d)", len(*tenants))
	}
}

func TestHandleTenantsMalformedJSON(t *testing.T) {

	okResponse := `[
		{
			"ID":"5010b97c-1d4f-11ec-821e-4756fdf46484",
			 "AuthEventID":"55c8fb68-1d4f-11ec-aeb7-3b421a2838d0",
			 "TenantID":"87d9664c-1d4f-11ec-ba64-af1d5e634f86",
			 "TenantType":"ORGANISATION",
			 "TenantName":"My Glorious Organization",
			 "CreatedDateUTC":"2021-07-01T20:55:00.5717400",
			 "UpdatedDateUTC":"2021-09-21 19:44:35.5610020"
		 }
	]` // last date is missing a 'T'

	token := initToken()
	token.AccessToken = "hij"
	token.RefreshToken = "klm"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(okResponse))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	handler := token.HandleTenants

	req := httptest.NewRequest("GET", server.URL, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(resp.Body)

	if statusCode != 500 {
		t.Errorf("Status code %d != 500", statusCode)
	}
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Content type unexpected: %s\n", contentType)
	}

	if !strings.Contains(string(body), "parsing time") {
		t.Errorf("Expected 'parsing time' error, got %s", string(body))
	}
}
