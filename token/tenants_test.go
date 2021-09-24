package token

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// tenants example from the Xero documentation at
// https://developer.xero.com/documentation/guides/oauth2/auth-flow#5-check-the-tenants-youre-authorized-to-access
var tenantsString = `
[
    {
        "id": "e1eede29-f875-4a5d-8470-17f6a29a88b1",
        "authEventId": "d99ecdfe-391d-43d2-b834-17636ba90e8d",
        "tenantId": "70784a63-d24b-46a9-a4db-0e70a274b056",
        "tenantType": "ORGANISATION",
        "tenantName": "Maple Florist",
        "createdDateUtc": "2019-07-09T23:40:30.1833130",
        "updatedDateUtc": "2020-05-15T01:35:13.8491980"
    },
    {
        "id": "32587c85-a9b3-4306-ac30-b416e8f2c841",
        "authEventId": "d0ddcf81-f942-4f4d-b3c7-f98045204db4",
        "tenantId": "e0da6937-de07-4a14-adee-37abfac298ce",
        "tenantType": "ORGANISATION",
        "tenantName": "Adam Demo Company (NZ)",
        "createdDateUtc": "2020-03-23T02:24:22.2328510",
        "updatedDateUtc": "2020-05-13T09:43:40.7689720"
    },
    {
        "id": "74305bf3-12e0-45e2-8dc8-e3ec73e3b1f9",
        "authEventId": "d0ddcf81-f942-4f4d-b3c7-f98045204db4",
        "tenantId": "c3d5e782-2153-4cda-bdb4-cec791ceb90d",
        "tenantType": "PRACTICEMANAGER",
        "tenantName": null,
        "createdDateUtc": "2020-01-30T01:33:36.2717380",
        "updatedDateUtc": "2020-02-02T19:21:08.5739590"
    }
]
`

func TestTenantsDecode(t *testing.T) {
	token := initToken()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tenantsString))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	tenants, err := token.Tenants()
	if err != nil {
		t.Errorf("tenant extraction error : %s", err)
	}
	if len(*tenants) != 3 {
		t.Errorf("length of tenants want(3) got (%d)", len(*tenants))
	}
}

func TestTenantsDecodeHTTPFail(t *testing.T) {
	token := initToken()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(tenantsString))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	_, err := token.Tenants()
	if err == nil {
		t.Errorf("tenant extraction error : %s", err)
	}
}

func TestTenantsDecodeFail(t *testing.T) {
	token := initToken()

	tenantsString := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tenantsString))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	_, err := token.Tenants()
	if !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Errorf("unexpected error message: %s", err)
	}
	if err == nil {
		t.Errorf("tenant decode error : %s", err)
	}
}

func TestTenantsDecodeEmptyString(t *testing.T) {
	token := initToken()

	tenantsString = strings.ReplaceAll(tenantsString, "tenantId", "tenantIdentifier")
	fmt.Println(tenantsString)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tenantsString))
	}))
	defer server.Close()

	token.tenantURL = server.URL
	token.AccessToken = "abc"
	token.RefreshToken = "def"
	tt, err := token.Tenants()
	if err != nil {
		t.Errorf("Tenantid check -- unexected error: %s", err)
	}
	tenants := *tt
	if tenants[0].TenantID != "" {
		t.Errorf("Tenantid != \"\" : %s", tenants[0].TenantID)
	}
}
