package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Defaults
var (
	defaultTimeout = 5 * time.Second
	defaultRetries = 0
	defaultAgent   = "gohttp"
)

//
// OptionsStruct
//
type Options struct {
	useV6         bool          // Use only IPv6
	useV4         bool          // Use only IPv4
	timeout       time.Duration // connection timeout in seconds
	retries       int           // number of retries
	printstatus   bool          // Print status and TLS info
	printheader   bool          // Print HTTP headers
	printbody     bool          // Print body
	queryall      bool          // Query all server addresses
	sni           string        // Server Name Indication option
	cacert        string        // File containing PEM format CA certs
	clientcert    string        // File containing PEM format client cert
	clientkey     string        // File containing PEM format client key
	username      string        // Username
	password      string        // Password
	showcert      bool          // Show peer certificate
	showcertchain bool          // Show peer certificate chain
	noredirect    bool          // Don't follow redirects
	noverify      bool          // Don't verify server certificate
	useragent     string        // User-Agent string
}

// Options
var options = Options{
	useV6:         false,
	useV4:         false,
	timeout:       defaultTimeout,
	retries:       defaultRetries,
	printstatus:   true,
	printheader:   true,
	printbody:     false,
	queryall:      false,
	sni:           "",
	cacert:        "",
	clientcert:    "",
	clientkey:     "",
	username:      "",
	password:      "",
	showcert:      false,
	showcertchain: false,
	noverify:      false,
	useragent:     defaultAgent}

//
// doFlags - process command line options
//
func doFlags() string {

	var authbasic string

	help := flag.Bool("h", false, "print help string")
	flag.BoolVar(&options.useV6, "6", false, "use IPv6 only")
	flag.BoolVar(&options.useV4, "4", false, "use IPv4 only")
	flag.DurationVar(&options.timeout, "t", defaultTimeout, "query timeout")
	flag.BoolVar(&options.printstatus, "status", true, "print status and TLS info")
	flag.BoolVar(&options.printheader, "header", true, "print header")
	flag.BoolVar(&options.printbody, "body", false, "print body")
	flag.BoolVar(&options.queryall, "queryall", false, "query all server addresses")
	flag.BoolVar(&options.noredirect, "noredirect", false, "don't follow redirects")
	flag.StringVar(&options.sni, "sni", "", "Server Name Indication")
	flag.StringVar(&options.cacert, "cacert", "", "CA cert file")
	flag.StringVar(&options.clientcert, "clientcert", "", "Client cert file")
	flag.StringVar(&options.clientkey, "clientkey", "", "Client key file")
	flag.StringVar(&authbasic, "authbasic", "", "Basic auth username:password")
	flag.BoolVar(&options.showcert, "showcert", false, "Show peer certificate")
	flag.BoolVar(&options.showcertchain, "showcertchain", false, "Show peer certificate chain")
	flag.BoolVar(&options.noverify, "noverify", false, "Don't verify server certificate")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s, version %s
Usage: %s [Options] <url>

    Options:
	-h                Print this help string
	-4                Connect to IPv4 addresses only
	-6                Connect to IPv6 addresses only
	-t Ns             Query timeout value in seconds (default %v)
	-r N              Maximum # of retries (default %d)
	-status           Print status and TLS info (=false to negate)
	-header           Print HTTP headers (=false to negate)
	-body             Print body
	-queryall         Query all server addresses (implies 'noredirect')
	-noredirect       Don't follow redirects
	-sni name         Server Name Indication option
	-cacert file      PEM format CA certificates file
	-clientcert file  PEM format Client certificate file
	-clientkey file   PEM format Client key file
	-authbasic creds  username:password string for basic authentication
	-showcert         Show peer certificate
	-showcertchain    Show peer certificate chain
	-noverify         Don't verify server certificate
`, progname, Version, progname, defaultTimeout, defaultRetries)
	}

	flag.Parse()

	if authbasic != "" {
		tmp := strings.SplitN(authbasic, ":", 2)
		options.username = tmp[0]
		options.password = tmp[1]
	}

	if options.queryall {
		options.noredirect = true
	}

	if *help || (flag.NArg() != 1) {
		if flag.NArg() != 0 {
			fmt.Fprintf(os.Stderr, "Error: incorrect number of arguments\n")
		}
		flag.Usage()
		os.Exit(4)
	}

	if options.useV4 && options.useV6 {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both -4 and -6.\n")
		flag.Usage()
		os.Exit(4)
	}

	return flag.Args()[0]
}
