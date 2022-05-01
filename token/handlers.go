package token

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// HandleLogin provides a login page if the client credentials (clientid
// and secret) together with the tenantid have not been set. If they
// have been, redirect to the home page
func (t *Token) HandleLogin(w http.ResponseWriter, r *http.Request) {

	if t.clientLoggedIn {
		// redirect to the /home endpoint
		w.Header().Set("Location", "/home")
		w.WriteHeader(302)
		return
	}

	var errorMsg string
	if r.Method == http.MethodPost {
		err := t.AddClientCredentials(
			r.PostFormValue("client"),
			r.PostFormValue("secret"),
			r.PostFormValue("tenantid"),
		)
		if err == nil {
			w.Header().Set("Location", "/home")
			w.WriteHeader(302)
			return
		}
		errorMsg = err.Error()
	}

	tpl := template.New("inline")
	tmpl, err := tpl.Parse(`
	<html><title>XeroOauthTokenServer login</title>
	<style>
	p.error { color: red }
	body { margin: 5% }
	label { display: inline-block; margin-bottom: 4px; width: 120px }
	</style>
	<body>
	<h3>XeroOauthTokenServer Login</h3>
	<p>Use this form to proceed to the next stage of login via Xero.</p>
	<h4>Xero client credentials</h4>
	{{ if . }}<p class="error">Error: {{ . }}{{ end }}
    <form method="POST">
        <label>ClientID:</label>
        <input size=32 type="text" name="client"><br />
        <label>Secret:</label>
        <input size=48 type="text" name="secret"><br />
        <label>TenantID:</label>
        <input size=32 type="text" name="tenantid"><br />
        <input type="submit">
    </form>
	</body>
	</html>
	`)
	if err != nil {
		log.Printf("form error: %s", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
	}
	tmpl.Execute(w, errorMsg)
}

// HandleHome provides the home page
func (t *Token) HandleHome(w http.ResponseWriter, r *http.Request) {

	if !t.clientLoggedIn {
		// redirect to the /login endpoint
		w.Header().Set("Location", "/login")
		w.WriteHeader(302)
		return
	}

	tpl := template.New("inline")
	tmpl, err := tpl.Parse(`
	<html><title>XeroOauthTokenServer : Xero login</title>
	<style>
	body { margin: 5% }
	label { display: inline-block; margin-bottom: 4px; width: 120px }
	</style>
	<body>
	<h3>Xero Login</h3>
	<p>As you have now provided the client credentials, you can proceed
	to the next stage of logging in with Xero</p>
	{{if .AccessToken }}
		<h4>Server initialised</h4>
		<p>The server is already initialised. However you can re-login using the
		code generation link below.</p>
		<p>View or extract the server token, refresh token and other details at the
		<a href="/status">/status</a> json endpoint.</p>
		<p>View or extract the current token at <a href="/token">/token</a></p>
		<p>Force a refresh at <a href="/refresh">/refresh</a></p>
		<p>Revoke a token using <a href="/refresh">/logout</a></p>
		<p>Logout and revoke the token using <a href="/refresh">/logout</a></p>
	{{else}}
		<h4>Code generation</h4>
		<p>Generate a code by <a href={{ .AuthURL }}>logging into Xero</a></p>
		<p>The code will then be swapped for a token and refresh token.</p>
	{{end}}
	</body></html>
	`)
	if err != nil {
		log.Printf("home page template error: %s", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
	tmpl.Execute(w, t)
}

// HandleCode is the code endpoint processes the code received from Xero
// to receive a token. The "code" response redirects here. Note that if
// "code" is provided the "state" string should be checked against the
// randomised string stored in the token struct; this is a security
// measure to avoid spoofed callouts.
func (t *Token) HandleCode(w http.ResponseWriter, r *http.Request) {

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusForbidden)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		msg := fmt.Sprint("No code to extract")
		log.Println(msg)
		http.Error(w, msg, http.StatusForbidden)
		return
	}

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

	err := t.GetToken(strings.TrimSpace(code))
	if err != nil {
		e, ok := err.(*HTTPClientError)
		var msg string
		if ok {
			msg = fmt.Sprintf("token retrieval failed: %d : %s", e.code, e.message)
		} else {
			msg = fmt.Sprintf("token retrieval error: %s", err)
		}
		log.Println(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	fmt.Fprint(w, "<html><title>Code extraction</title><body>")
	fmt.Fprint(w, "<h4>Code extraction succeeded</h4>")
	fmt.Fprint(w, `<p>View the <a href="/token">token</a>, `)
	fmt.Fprint(w, `<a href="/refresh">refresh the token</a> `)
	fmt.Fprint(w, `or view the service <a href="/status">status</a>.</p>`)
}

// HandleLivez checks if the application is healthy
func (t *Token) HandleLivez(w http.ResponseWriter, r *http.Request) {
	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleStatus shows the status of the server/tokenserver struct
func (t *Token) HandleStatus(w http.ResponseWriter, r *http.Request) {

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	j, err := t.AsJSON()
	if err != nil {
		msg := fmt.Sprintf("status json encoding error: %s", err)
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

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	n := time.Now()
	err := t.Refresh()
	if err != nil {
		e, ok := err.(*HTTPClientError)
		var msg string
		if ok {
			msg = fmt.Sprintf("refresh failed: %d : %s", e.code, e.message)
		} else {
			msg = fmt.Sprintf("refresh error: %s", err)
		}
		log.Println(msg)
		http.Error(w, msg, http.StatusNotFound)
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

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

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

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

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

// HandleTenants returns the tenants accessible with this token. If the
// token has expired refresh has to be handled manually
func (t *Token) HandleTenants(w http.ResponseWriter, r *http.Request) {

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	tenants, err := t.Tenants()
	if err != nil {
		msg := fmt.Sprintf("tenant retrieval error: %s", err)
		msg = msg + "\nyou may need to run /refresh"
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(tenants)
	if err != nil {
		msg := fmt.Sprintf("tenant encoding error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// HandleRevoke runs the revocation function for revoking a token and
// all of its connections
func (t *Token) HandleRevoke(w http.ResponseWriter, r *http.Request) {

	if !t.clientLoggedIn {
		msg := "client has not logged in"
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if t.AccessToken == "" || t.RefreshToken == "" {
		msg := "system has not been initialised or is in an error state"
		log.Println(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
	err := t.Revoke()
	if err != nil {
		msg := fmt.Sprintf("error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(map[string]string{"status": "revoked"})
	w.Write(j)
}

// HandleLogout runs the revocation function and removes the login
// details; this is a client facing call
func (t *Token) HandleLogout(w http.ResponseWriter, r *http.Request) {

	// ignore errors for revocation and client credentials clearing
	t.Revoke()
	t.Logout()

	w.Header().Set("Location", "/")
	w.WriteHeader(302)
}
