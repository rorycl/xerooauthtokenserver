package token

import (
	"testing"
	"time"
)

func TestRefresher(t *testing.T) {

	now := time.Now().UTC()

	token.expireTimeTicker = 30 * time.Millisecond
	token.AccessTokenExpiryUTC = now.Add(200 * time.Millisecond)
	token.accessTokenExpirySecs = 100 * time.Millisecond
	token.RefreshTokenExpiryUTC = now.Add(300 * time.Millisecond)
	token.refreshTokenExpirySecs = 100 * time.Millisecond

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
