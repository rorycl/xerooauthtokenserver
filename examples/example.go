/*
Example showing basic usage of github.com/rorycl/XeroOauthTokenServer/token
MIT licence
Rory Campbell-Lange 08 September 2021

This requires an app configured in Xero at
https://developer.xero.com/app/manage/app

The programme redirects to xero.com from which the code element needs to
be extracted and pasted into the command line.

It is important that "https://xero.com/" is one of the redirect domains
permitted in your OAuth 2.0 credentials "redirect URIs" set on the Xero
app configuration page.
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rorycl/XeroOauthTokenServer/token"
)

func main() {

	redirect := "https://xero.com/"
	clientID := os.Getenv("XEROCLIENTID")
	clientSecret := os.Getenv("XEROCLIENTSECRET")
	scopes := []string{"offline_access", "accounting.transactions"}

	if clientID == "" || clientSecret == "" {
		fmt.Println("Please set the client id and secret environmental variables")
		os.Exit(1)
	}

	authURL, tokenURL := "", "" // use defaults
	ts, err := token.NewToken(redirect, clientID, clientSecret, scopes, authURL, tokenURL)
	if err != nil {
		fmt.Printf("new tokenServer error %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Please go to the url below and log into Xero")
	fmt.Println(ts.AuthURL())
	fmt.Println("")

	fmt.Printf("Please paste the 'code' component of the url here:\n")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	code := scanner.Text()

	fmt.Println("\nShowing code:")
	fmt.Println(code)

	fmt.Println("\nGetting token:")
	ts.GetToken(strings.TrimSpace(code))
	fmt.Println(ts)

	time.Sleep(2 * time.Second)

	fmt.Println("\nRefreshing token:")
	ts.Refresh()
	fmt.Println(ts)

}
