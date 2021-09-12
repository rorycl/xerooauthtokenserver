# xerooauthtokenserver

version 0.0.5 September 2021

## Summary

xerooauthtokenserver is an http server for managing OAuth2 tokens for
the accounting software as a service system Xero. The server acts as a
sidecar or proxy, providing an easy upgrade path for software designed
for the previous Xero OAuth1 flow.

After following the Xero Oauth2 flow the server makes available a json
access token at the /token endpoint, and will refresh tokens when
required using the refresh token, or when the refresh token is due to
expire.

xerooauthtokenserver/token also provides a package to easily integrate
the Xero OAuth2 flow into a Go programme.

## Usage

Ensure you have configurated a Xero OAuth2 web app at
`https://developer.xero.com/app/manage` and generated a client id and
secret. It is important that one of the "Redirect URIs" is set to
`http://localhost:5001/` or your locally configured server address.

Run the server (you may want to set the `XEROCLIENTID` and
`XEROCLIENTSECRET` environmental variables for convenience first) and
navigate to `http://127.0.0.1:5001` and follow the Xero authentication
flow. You can then extract a token at the `/token` endpoint, force a
refresh at `/refresh` or view the server data at `/healthz`.

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
./XeroOauthTokenServer -i CLIENTID -s CLIENTSECRET \
                       --refreshtoken=<refreshtokendata>
```

Security: it is not advisable to put this server on the public internet.

## Programme options

```
Usage:
  XeroOauthTokenServer  <options>

  Xero oauth token server : 0.0.5 September 2021


Application Options:
  -p, --port=         port to run on (default: 5001)
  -n, --address=      network address to run on (default: 127.0.0.1)
  -i, --clientid=     xero client id, or use env [$XEROCLIENTID]
  -s, --clientsecret= xero client secret, or use env [$XEROCLIENTSECRET]
  -r, --redirect=     oauth2 redirect address (default: http://localhost:5001/)
  -o, --scopes=       oauth2 scopes (default: offline_access, accounting.transactions)
  -m, --refreshmins=  set lifetime of refresh token (default 50 days) (default: 4320000)
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
    """retrieve first tenant id"""
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
