/*
xerooauthtokenserver v0.0.7

https://github.com/rorycl/xerooauthtokenserver

Summary:

XeroOauthTokenServer is an http server for managing OAuth2 tokens for
the accounting software as a service system Xero. The server acts as a
sidecar or proxy.

After taking the user through first a login screen for providing the
Xero client id, client secret and tenant id, the programme then takes
the user through the Xero Oauth2 flow. If this is successful, the server
makes a json access token available at the /token endpoint, and will
refresh tokens when required, or when the refresh token is due to
expire.

The `xerooauthtokenserver/token` package makes it easy to integrate the
Xero OAuth2 flow into a Go programme, including automated token
refreshes.

It is not advisable to put this server on the public internet.

This server is only suitable for integration with software designed to
act with the permissions of the user initialising it.

This software is provided under an MIT licence.
*/

package main
