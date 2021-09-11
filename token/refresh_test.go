package token

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefresher(t *testing.T) {

	now := time.Now().UTC()

	token.expireTimeTicker = 30 * time.Millisecond
	token.RefreshTokenExpiryUTC = now.Add(200 * time.Millisecond)
	token.expirySecs = 100 * time.Millisecond

	after := time.After(120 * time.Millisecond)

	refresher := token.refresher()

	counter := 0
loop:
	for {
		select {
		case <-refresher:
			counter++
			if counter > 2 {
				break loop
			}
		case <-after:
			t.Errorf("Timeout triggered")
			break loop
		}
	}
	if counter != 3 {
		t.Errorf("Expect 3 ticks during test, got %d", counter)
	}
}

func TestTriggerRefreshRunner(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "abc", "refresh_token": "def", "expires_in": 1800}`))
	}))
	defer server.Close()

	token.tokenURL = server.URL
	token.AccessToken = "ghi"
	token.RefreshToken = "jkl"
	err := token.Refresh()
	if err != nil {
		t.Errorf("token.Refresh returned error %s", err)
	}

	refresher := make(chan struct{})
	token.refreshRunner(refresher)
	refresher <- struct{}{}

	if token.AccessToken != "abc" {
		t.Errorf("access token error have(%s) want(%s)", token.AccessToken, "abc")
	}
	if token.RefreshToken != "def" {
		t.Errorf("refresh token error have(%s) want(%s)", token.RefreshToken, "def")
	}

}
