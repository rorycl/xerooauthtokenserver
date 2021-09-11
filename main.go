package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/braintree/manners"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	flags "github.com/jessevdk/go-flags"
	"github.com/rorycl/XeroOauthTokenServer/token"
)

const description = "Xero oauth token server"
const version = "0.0.3 September 2021"
const usage = " <options>" + "\n\n  " + description

// Opts are the command line options
type Opts struct {
	Port         string   `short:"p" long:"port" description:"port to run on" default:"5001"`
	Addr         string   `short:"n" long:"address" description:"network address to run on" default:"127.0.0.1"`
	ClientID     string   `short:"i" long:"clientid" description:"xero client id, or use env" env:"XEROCLIENTID" required:"yes"`
	ClientSecret string   `short:"s" long:"clientsecret" description:"xero client secret, or use env" env:"XEROCLIENTSECRET" required:"yes"`
	Redirect     string   `short:"r" long:"redirect" description:"oauth2 redirect address" default:"http://localhost:5001/"`
	Scopes       []string `short:"o" long:"scopes" description:"oauth2 scopes" default:"offline_access" default:"accounting.transactions"`
	RefreshMins  int      `short:"m" long:"refreshmins" description:"refresh token within this number of minutes (default 20 days)" default:"28800"`
}

func main() {

	var options Opts
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = fmt.Sprintf("%s : %s\n", usage, version)

	if _, err := parser.Parse(); err != nil {
		flagError := err.(*flags.Error)
		if flagError.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
		}
		os.Exit(1)
	}

	if options.RefreshMins < 20 {
		log.Printf("It is inadvisable to set the refresh interval to less than 20 minutes in production")
	}

	authURL, tokenURL := "", "" // use Xero default urls
	ts, err := token.NewToken(
		options.Redirect,
		options.ClientID,
		options.ClientSecret,
		options.Scopes,
		authURL,
		tokenURL,
	)

	if err != nil {
		log.Printf("new token server error %s\n", err)
		os.Exit(1)
	}

	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()
	r.HandleFunc("/", ts.HandleHome)
	r.HandleFunc("/code", ts.HandleCode)
	r.HandleFunc("/healthz", ts.HandleHealthz)
	r.HandleFunc("/token", ts.HandleAccessToken)
	r.HandleFunc("/refresh", ts.HandleRefresh)

	// create a handler wrapped in a recovery handler and logging handler
	hdl := handlers.RecoveryHandler()(
		handlers.LoggingHandler(os.Stdout, r))

	// configure server options
	server := &http.Server{
		Addr:         options.Addr + ":" + options.Port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      hdl,
	}

	// wrap server with manners
	manners.ListenAndServe(options.Addr+":"+options.Port, server.Handler)

	log.Printf("serving on %s:%s", options.Addr, options.Port)

	// catch signals
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go listenForShutdown(ch)

}

func listenForShutdown(ch <-chan os.Signal) {
	<-ch
	log.Print("Closing the Rtest http server")
	manners.Close()
}
