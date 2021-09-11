package token

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// HandleHome provides the home page; the "code" response redirects
// here, so redirect to the code endpoint if so. Note that if "code" is
// provided the "state" string should be checked against the randomised
// string stored in the token struct; this is a security measure to
// avoid spoofed callouts.
func (t *Token) HandleHome(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		state := r.URL.Query().Get("state")
		if state != t.state {
			msg := fmt.Sprintf(
				"url state != saved state: %s %s",
				r.URL.RawQuery, t.state,
			)
			log.Println(msg)
			http.Error(w, msg, http.StatusForbidden)
			return
		}
		// redirect to the /code endpoint
		w.Header().Set("Location", fmt.Sprintf("/code?code=%s", code))
		w.WriteHeader(302)
		return
	}

	fmt.Fprint(w, "<html><title>Xero login</title><body>")
	fmt.Fprint(w, "<h4>Code generation</h4>")
	fmt.Fprintf(w, "<p>Generate a code by <a href=\"%s\">logging into Xero</a></p>",
		t.AuthURL())
	fmt.Fprint(w, "<p>The code will then be swapped for a token and refresh token.</p>")
	fmt.Fprint(w, "</body></html>")
}

// HandleCode is the code endpoint processes the code received from Xero
// to receive a token
func (t *Token) HandleCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		msg := fmt.Sprint("No code to extract")
		log.Println(msg)
		http.Error(w, msg, http.StatusForbidden)
		return
	}
	err := t.GetToken(strings.TrimSpace(code))
	if err != nil {
		msg := fmt.Sprintf("token retrieval error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusServiceUnavailable)
		return
	}
	fmt.Fprint(w, "<html><title>Code extraction</title><body>")
	fmt.Fprint(w, "<h4>Code extraction succeeded</h4>")
	fmt.Fprint(w, `<p>View the <a href="/token">token</a>, `)
	fmt.Fprint(w, `<a href="/refresh">refresh the token</a> `)
	fmt.Fprint(w, `or view the service <a href="/healthz">health</a>.</p>`)
}

// HandleHealthz shows the status of the server/tokenserver struct
func (t *Token) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	j, err := t.AsJSON()
	if err != nil {
		msg := fmt.Sprintf("healthz json encoding error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

// HandleRefresh handles refreshing a token, redirecting to the /token
// endpoint if successful
func (t *Token) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	n := time.Now()
	err := t.Refresh()
	if err != nil {
		msg := fmt.Sprintf("refresh error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusServiceUnavailable)
		return
	}
	a := time.Since(n)
	log.Printf("Refresh took: %s\n", a)
	w.Header().Set("Location", "/token")
	w.WriteHeader(302)
	return
}

// HandleAccessToken returns a json token
func (t *Token) HandleAccessToken(w http.ResponseWriter, r *http.Request) {
	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	// Get or refresh the token
	_, err := t.Get()
	if err != nil {
		msg := fmt.Sprintf("token get or refresh error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// jsonify
	j, err := t.TokenJSON()
	if err != nil {
		msg := fmt.Sprintf("token json encoding error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

// HandleRefreshToken returns a json refresh token
func (t *Token) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	j, err := t.RefreshTokenJSON()
	if err != nil {
		msg := fmt.Sprintf("refresh token json encoding error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
