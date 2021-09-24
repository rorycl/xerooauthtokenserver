package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// jsonDateTime is for encoding/decoding json date formats
type jsonDateTime struct {
	time.Time
}

const jsonDateTimeFMT = "2006-01-02T15:04:05.0000000"

// UnmarshalJSON unmarshals from a RFC3339 format date (without the "Z")
func (t *jsonDateTime) UnmarshalJSON(buf []byte) error {
	tt, err := time.Parse(jsonDateTimeFMT, strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

// MarshalJSON marshals a jsonDateTime to a RFC3339 format date string (without the "Z")
func (t jsonDateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format(jsonDateTimeFMT) + `"`), nil
}

// Tenants represents a slice of Xero tenants
type Tenants []struct {
	ID             string       `json:"id"`
	AuthEventID    string       `json:"authEventId"`
	TenantID       string       `json:"tenantId"`
	TenantType     string       `json:"tenantType"`
	TenantName     string       `json:"tenantName"`
	CreatedDateUTC jsonDateTime `json:"createdDateUtc"`
	UpdatedDateUTC jsonDateTime `json:"updatedDateUtc"`
}

// Tenants retrieves Xero tenants
func (t *Token) Tenants() (tenants *Tenants, err error) {

	req, err := http.NewRequest("GET", t.tenantURL, nil)
	if err != nil {
		return tenants, err
	}
	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: t.httpclientTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return tenants, err
	}
	if resp.StatusCode != 200 {
		return tenants, fmt.Errorf(
			"Tenant callout http error, %d",
			resp.StatusCode,
		)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tenants, fmt.Errorf(
			"Tenant callout failed, body read error, %s",
			string(body),
		)
	}
	err = json.Unmarshal(body, &tenants)
	return tenants, err
}
