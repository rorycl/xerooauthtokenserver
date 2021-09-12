/*
xerooauthtokenserver v0.0.5

https://github.com/rorycl/xerooauthtokenserver

Summary:

XeroOauthTokenServer is an http server for managing OAuth2 tokens for
the accounting software as a service system Xero. The server acts as a
sidecar or proxy, providing an easy upgrade path for software designed
for the previous Xero OAuth1 flow.

After following the Xero Oauth2 flow the server makes available a json
access token at the /token endpoint, and will refresh tokens when
required using the refresh token, or when the refresh token is due to
expire.

The server can also be initialised with a save refresh token.

The xerooauthtokenserver/token package provides a convenient way to
integrate Xero Oauth2 flows into a Go programme.

This software is provided under an MIT licence, with no warranty.
*/

package main
