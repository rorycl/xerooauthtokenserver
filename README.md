# xerooauthtokenserver

version 0.0.7 October 2021

## Summary

XeroOauthTokenServer is an http server for managing OAuth2 tokens for
the Xero accounting SaaS service. The server acts as a sidecar or proxy,
providing an easy upgrade path for software designed for the previous
Xero OAuth1 flow or for avoiding having to integrate OAuth2 in-service.

After taking the user through the Xero Oauth2 flow the server makes a
json access token available at the /token endpoint, and will refresh
tokens when required, or when the refresh token is due to expire.

The `xerooauthtokenserver/token` package makes it easy to integrate the
Xero OAuth2 flow into a Go programme, including automated token
refreshes.

Please note the security warnings set out in `Security and Warranty`
below.

## Usage

Ensure you have configured a Xero OAuth2 web app at
https://developer.xero.com/app/manage and generated a client id and
secret. It is important that one of the "Redirect URIs" is set to
`http://localhost:5001/` or your locally configured XeroOauthTokenServer
server address if it is not at `127.0.0.1:5001`.

Run the server (you may want to set the `XEROCLIENTID` and
`XEROCLIENTSECRET` environmental variables for convenience first) and
navigate to `http://127.0.0.1:5001` and follow the Xero authentication
flow. You can then extract a token at the `/token` endpoint, force a
refresh at `/refresh` or view the server data at `/livez`.

For example, invoke the programme as follows:

```bash
./XeroOauthTokenServer -i CLIENTID -s CLIENTSECRET
```

or with the `XEROCLIENTID` and `XEROCLIENTSECRET` environmental
variables set:

```bash
./XeroOauthTokenServer
```

If you have previously saved or extracted a Xero refresh token, perhaps
with the Xero [XOAuth](https://github.com/XeroAPI/xoauth) tool, you can
initialise the server and trigger an immediate token retrieval by
invoking the server as follows:

```bash
./XeroOauthTokenServer -i CLIENTID -s CLIENTSECRET --refreshtoken=REFRESHTOKEN
```

It is necessary to revoke a token (using the associated refresh token)
to limit or expand scopes. If a different set of scopes is specified
to that associated with a refresh token the programme will abort with an
error.

The following endpoints are provided:

```
/        : home - use this to commence the oauth2 flow
/code    : used as the redirect endpoint
/livez   : check service health
/status  : view the status of the services
/token   : view the current token
/refresh : force a refresh of the token
/tenants : view the tenants accessible with this token
/revoke  : revoke the token
```

## Security and Warranty

It is not advisable to put this server on the public internet.

This server is only suitable for integration with software designed to
act with the permissions of the user initialising it.

Please refer to the LICENSE and note that this software is provided
without warranty of any kind.

## Programme options

```
Usage:
  XeroOauthTokenServer  <options>

  Xero oauth token server : 0.0.7 September 2021

Application Options:
  -p, --port=         port to run on (default: 5001)
  -n, --address=      network address to run on (default: 127.0.0.1)
  -i, --clientid=     xero client id, or use env [$XEROCLIENTID]
  -s, --clientsecret= xero client secret, or use env [$XEROCLIENTSECRET]
  -r, --redirect=     oauth2 redirect address (default: http://localhost:5001/)
  -o, --scopes=       oauth2 scopes (default: offline_access, accounting.transactions, accounting.reports.read)
  -m, --refreshmins=  set lifetime of refresh token (default 50 days) (default: 72000)
      --refreshtoken= initialize server with refresh token

Help Options:
  -h, --help          Show this help message

```

## Integration

An integration example is provided in the `examples` directory of this
repo, and reproduced below. The integration assumes xerooauthtokenserver
is running on the default address and port. The integration is
encapsulated by the `get_token` function.

```python
"""
Python xerooauthtokenserver integration example
"""

import requests


def get_token():
    """retrieve token from xerooauthtoken server"""
    response = requests.get("http://127.0.0.1:5001/token")
    return response.json()['accessToken']


def tenants(access_token):
    """
    retrieve the id of the first tenant
    can also use "http://127.0.0.1:5001/tenants"
    """
    tenants_url = 'https://api.xero.com/connections'
    response = requests.get(
        tenants_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Content-Type': 'application/json'
        })
    return response.json()[0]["tenantId"]  # first tenant id


def invoices(access_token, tenant_id):
    """get invoices"""
    invoice_url = 'https://api.xero.com/api.xro/2.0/Invoices'
    response = requests.get(
        invoice_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Xero-tenant-id': tenant_id,
            'Accept': 'application/json'
        })
    return response.json()


if __name__ == '__main__':

    token = get_token()
    tenant = tenants(token)
    invoices = invoices(token, tenant)
    print(invoices)

```
