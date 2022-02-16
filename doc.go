/*
xerooauthtokenserver v0.0.7

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

The server can also be initialised with a saved refresh token.

The xerooauthtokenserver/token package provides a convenient way to
integrate Xero Oauth2 flows into a Go programme.

It is not advisable to put this server on the public internet.

This server is only suitable for integration with software designed to
act with the permissions of the user initialising it.

This software is provided under an MIT licence.
*/

package main
