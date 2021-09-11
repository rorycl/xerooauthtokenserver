/*
xerooauthtokenserver v0.0.4

Summary:

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
*/

package main
