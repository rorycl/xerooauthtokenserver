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
	"log"
	"os"
	"strings"
	"time"

	"github.com/rorycl/XeroOauthTokenServer/token"
)

func main() {

	redirect := "https://xero.com/"
	clientID := os.Getenv("XEROCLIENTID")
	clientSecret := os.Getenv("XEROCLIENTSECRET")
	tenantID := os.Getenv("XEROTENANTID")
	scopes := []string{"offline_access", "accounting.transactions"}

	if clientID == "" || clientSecret == "" || tenantID == "" {
		fmt.Println("Please set the client id, secret and tenant environmental variables")
		fmt.Println("XEROCLIENTID XEROCLIENTSECRET and XEROTENANTID")
		os.Exit(1)
	}

	authURL, tokenURL, tenantURL, refreshLifetime := "", "", "", 0 // use defaults
	ts, err := token.NewToken(redirect, scopes, authURL, tokenURL, tenantURL, refreshLifetime)
	if err != nil {
		fmt.Printf("new tokenServer error %s\n", err)
		os.Exit(1)
	}
	err = ts.AddClientCredentials(clientID, clientSecret, tenantID)
	if err != nil {
		fmt.Printf("could not add credentials %s\n", err)
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
	err = ts.GetToken(strings.TrimSpace(code))
	if err != nil {
		log.Fatalf("tenants error: %s\n", err)
	}

	fmt.Println(ts)

	fmt.Println("\nGetting tenants:")
	tenants, err := ts.Tenants()
	if err != nil {
		log.Fatalf("tenants error: %s\n", err)
	}
	fmt.Printf("%v\n", tenants)

	time.Sleep(2 * time.Second)

	fmt.Println("\nRefreshing token:")
	ts.Refresh()
	fmt.Println(ts)

	time.Sleep(2 * time.Second)

	fmt.Println("\nGetting tenants:")
	tenants, err = ts.Tenants()
	if err != nil {
		log.Fatalf("tenants error: %s\n", err)
	}
	fmt.Printf("%v\n", tenants)

}
