package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rorycl/XeroOauthTokenServer/randstring"
)

// XeroAuthURL is the Xero authorization url
const XeroAuthURL string = "https://login.xero.com/identity/connect/authorize"

// XeroTokenURL is the Xero token receipt url
const XeroTokenURL string = "https://identity.xero.com/connect/token"

// XeroTenantURL is the Xero tenant endpoint
const XeroTenantURL = "https://api.xero.com/connections"

// XeroRefreshExpirationDays is the default expiration of the refresh
// token from now, which is 60 days; let's say 50
// See https://developer.xero.com/faq/oauth2/
const XeroRefreshExpirationDays int = 50

// DefaultExpirySecs is the number of seconds before the any token
// expiry to trigger a refresh, for instance <n> seconds before the
// refresh token expiry
const DefaultExpirySecs int = 60

// Token represents Xero API Tokens provided by the Xero OAuth2 flow,
// particularly each AccessToken which is valid for 30 minutes and
// RefreshTokens which are valid for 30 days. The tokens are also scoped
// by Scopes.

// The private identifiers redirectURL, clientID and clientSecret are
// used for initial authentication which, together with a randomized
// "state" identifier, returns a code which is exchanged for an access
// token and refresh token.

// The randomised state string is written to the Token struct in a way
// that could cause a race condition if the Auth.URL call is made twice
// before the code exchange for the first url is completed.

// The Token data structure is locked via a sync.Mutex on update.
type Token struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiryUTC  time.Time `json:"access_token_expiry_utc"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiryUTC time.Time `json:"refresh_token_expiry_utc"`
	Scopes                []string  `json:"scopes"`
	clientID              string
	clientSecret          string
	state                 string
	authURL               string
	redirectURL           string
	tokenURL              string
	tenantURL             string
	httpclientTimeout     time.Duration
	expireTimeTicker      time.Duration
	expirySecs            time.Duration
	refreshTokenLifetime  time.Duration
	locker                sync.Mutex
	refreshChan           <-chan struct{}
}

// String represents Token for printing
func (t *Token) String() string {
	tpl := `
access_token   %s
expiry         %s
refresh_token  %s
refresh_expiry %s
scopes         %v
`
	return fmt.Sprintf(
		tpl,
		t.AccessToken,
		t.AccessTokenExpiryUTC,
		t.RefreshToken,
		t.RefreshTokenExpiryUTC,
		t.Scopes,
	)
}

// AsJSON returns a json encoding for a Tokenserver
func (t *Token) AsJSON() (j []byte, err error) {
	return json.Marshal(t)
}

// TokenJSON returns a json respresentation of a token
func (t *Token) TokenJSON() (j []byte, err error) {
	ts := map[string]string{"accessToken": t.AccessToken}
	return json.Marshal(ts)
}

// RefreshTokenJSON returns a json respresentation of a refresh token
func (t *Token) RefreshTokenJSON() (j []byte, err error) {
	ts := map[string]string{"refreshToken": t.RefreshToken}
	return json.Marshal(ts)
}

// NewToken returns a new Token struct
func NewToken(redirect, client, secret string, scopes []string, authURL, tokenURL, tenantURL string, refreshMins int) (t *Token, err error) {

	_, err = url.ParseRequestURI(redirect)
	if err != nil {
		return t, errors.New("redirect url invalid")
	}
	if client == "" || secret == "" {
		return t, errors.New("redirect, client or secret is empty")
	}
	if authURL == "" {
		authURL = XeroAuthURL
	}
	if tokenURL == "" {
		tokenURL = XeroTokenURL
	}
	if tenantURL == "" {
		tenantURL = XeroTenantURL
	}
	if len(scopes) < 1 {
		return t, errors.New("scopes cannot be empty")
	}

	var refreshLifetime time.Duration
	if refreshMins == 0 {
		refreshLifetime = time.Hour * time.Duration(24*XeroRefreshExpirationDays)
	} else {
		refreshLifetime = time.Minute * time.Duration(refreshMins)
	}

	t = &Token{
		redirectURL:          redirect,
		clientID:             client,
		clientSecret:         secret,
		Scopes:               scopes,
		authURL:              authURL,
		tokenURL:             tokenURL,
		tenantURL:            tenantURL,
		httpclientTimeout:    time.Second * 3,
		expireTimeTicker:     time.Minute * 1,
		expirySecs:           time.Second * time.Duration(DefaultExpirySecs),
		refreshTokenLifetime: refreshLifetime,
	}

	// initialise goroutines for refreshing tokens
	t.refreshChan = t.refresher()
	t.refreshRunner(t.refreshChan)

	return t, nil
}

// AuthURL returns the authorization url which is the beginning of the
// authorization process; the state string is randomized and stored in t
// (note that this could cause a race condition)
func (t *Token) AuthURL() string {

	t.state = randstring.RandString(10)

	scope := ""
	for _, s := range t.Scopes {
		scope += fmt.Sprintf(" %s", s)
	}
	scope = url.QueryEscape(strings.TrimSpace(scope))

	// todo: move to url.URL
	tpl := t.authURL + "?" + "response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s"
	url := fmt.Sprintf(tpl, t.clientID, t.redirectURL, scope, t.state)

	return url
}

// encodeIDSecret encodes the clientid and clientsecret into a "basic"
// string suitable for an authentication header
func (t *Token) encodeIDSecret() string {
	s := fmt.Sprintf(
		"Basic %s:%s",
		t.clientID,
		t.clientSecret,
	)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// setExpiry sets the UTC expiration time of the token and refreshtoken
func (t *Token) setExpiry(expiry int) {
	now := time.Now().UTC()
	t.AccessTokenExpiryUTC = now.Add(time.Duration(expiry) * time.Second)
	t.RefreshTokenExpiryUTC = now.Add(t.refreshTokenLifetime)
	log.Printf("Setting expiry: lifetime %v refresh %s", t.refreshTokenLifetime, t.RefreshTokenExpiryUTC)
}

// tokenResults is the type of the Xero API results
type tokenResults struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// GetToken retrieves a token if possible from an authorization code
func (t *Token) GetToken(code string) error {

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", t.redirectURL)
	req, err := http.NewRequest("POST", t.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", t.encodeIDSecret())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(t.clientID), url.QueryEscape(t.clientSecret))

	client := http.Client{
		Timeout: t.httpclientTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte("could not read body")
		}
		return &HTTPClientError{resp.StatusCode, string(body)}
	}

	var results tokenResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return fmt.Errorf("json decoding error: %s", err)
	}
	if results.AccessToken == "" || results.RefreshToken == "" || results.ExpiresIn == 0 {
		return errors.New("empty response received from server")
	}

	t.locker.Lock()
	t.AccessToken = results.AccessToken
	t.RefreshToken = results.RefreshToken
	t.setExpiry(results.ExpiresIn)

	t.locker.Unlock()

	return nil
}

// Refresh uses a refresh token to retrieve a new token and refresh token
func (t *Token) Refresh() error {

	if t.AccessToken == "" || t.RefreshToken == "" {
		return errors.New("token system has not been initialised")
	}

	// reset access token if the service is being refreshed from the
	// command line via a saved refresh token
	if t.AccessToken == "override" {
		t.AccessToken = ""
	}

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", t.RefreshToken)
	req, err := http.NewRequest("POST", t.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", t.encodeIDSecret())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(t.clientID), url.QueryEscape(t.clientSecret))

	client := http.Client{
		Timeout: t.httpclientTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte("could not read body")
		}
		return &HTTPClientError{resp.StatusCode, string(body)}
	}

	var results tokenResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return fmt.Errorf("json decoding error: %s", err)
	}
	if results.AccessToken == "" || results.RefreshToken == "" || results.ExpiresIn == 0 {
		return errors.New("empty response received from server")
	}

	t.locker.Lock()
	t.AccessToken = results.AccessToken
	t.RefreshToken = results.RefreshToken
	t.setExpiry(results.ExpiresIn)
	t.locker.Unlock()

	log.Printf("new refresh token registered: %s", t.RefreshToken)

	return nil
}

// Get returns the Token after refreshing if necessary. An assumption is
// made that some latitude (expirySecs) is needed when determining
// expiration.
func (t *Token) Get() (tt *Token, err error) {
	now := time.Now().UTC()
	if t.AccessTokenExpiryUTC.Add(-t.expirySecs).After(now) {
		return t, nil
	}
	log.Println("Running refresh")
	err = t.Refresh()
	return t, err
}
