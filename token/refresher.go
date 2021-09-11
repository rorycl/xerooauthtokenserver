package token

import (
	"log"
	"time"
)

// updater is a function that returns a channel to refresh a token if it
// is due to expire
func (t *Token) refresher() <-chan struct{} {
	ticker := time.NewTicker(t.expireTimeTicker)
	refresher := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if t.expiring() {
					refresher <- struct{}{}
				}
			}
		}
	}()
	return refresher
}

// expiring determines if either the Token or RefreshToken
func (t *Token) expiring() bool {
	now := time.Now().UTC()
	if t.RefreshTokenExpiryUTC.Add(-t.refreshTokenExpirySecs).After(now) {
		return true
	}
	return false
}

// triggerRefresh triggers a token refresh; separated from the refresher
// function to allow for testing
func (t *Token) triggerRefresh(refresher <-chan struct{}) {
	go func() {
		for range refresher {
			err := t.Refresh()
			if err != nil {
				log.Printf("refresh error %s", err)
			}
		}
	}()
}
