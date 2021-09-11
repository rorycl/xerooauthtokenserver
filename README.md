# xerooauthtokenserver

Local server to act as a Xero Oauth2 token proxy to help software
migrating from Xero's Oauth1 flows.

*Do not use* as this repo is still in development.

Run the server locally and use it to retrieve a token and refresh token
from Xero. The server will automatically refresh tokens and provide a
token at the `/token` endpoint.

## Usage

```
Usage:
  XeroOauthTokenServer  <options>

  Xero oauth token server : 0.0.3 September 2021

Application Options:
  -p, --port=         port to run on (default: 5001)
  -n, --address=      network address to run on (default: 127.0.0.1)
  -i, --clientid=     xero client id, or use env [$XEROCLIENTID]
  -s, --clientsecret= xero client secret, or use env [$XEROCLIENTSECRET]
  -r, --redirect=     oauth2 redirect address (default: http://localhost:5001/)
  -o, --scopes=       oauth2 scopes (default: offline_access, accounting.transactions)
  -m, --refreshmins=  refresh token within this number of minutes (default 20 days) (default:
                      28800)

Help Options:
  -h, --help          Show this help message
```



