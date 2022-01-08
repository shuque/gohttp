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

type arrayFlag []string

func (i *arrayFlag) String() string {
	return "my string representation"
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

//
// OptionsStruct
//
type Options struct {
	ipv6only      bool          // Use only IPv6
	ipv4only      bool          // Use only IPv4
	timeout       time.Duration // connection timeout in seconds
	retries       int           // number of retries
	printbody     bool          // Print body
	bodyonly      bool          // Print body only
	queryall      bool          // Query all server addresses
	sni           string        // Server Name Indication option
	headers       arrayFlag     // Custom request headers
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
	ipv6only:      false,
	ipv4only:      false,
	timeout:       defaultTimeout,
	retries:       defaultRetries,
	printbody:     false,
	bodyonly:      false,
	queryall:      false,
	sni:           "",
	headers:       nil,
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
	flag.BoolVar(&options.ipv6only, "6", false, "use IPv6 only")
	flag.BoolVar(&options.ipv4only, "4", false, "use IPv4 only")
	flag.DurationVar(&options.timeout, "t", defaultTimeout, "query timeout")
	flag.BoolVar(&options.printbody, "printbody", false, "print body")
	flag.BoolVar(&options.bodyonly, "bodyonly", false, "print body")
	flag.BoolVar(&options.queryall, "queryall", false, "query all server addresses")
	flag.BoolVar(&options.noredirect, "noredirect", false, "don't follow redirects")
	flag.StringVar(&options.sni, "sni", "", "Server Name Indication")
	flag.Var(&options.headers, "header", "Custom request header: key: value")
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
	-4                Connect to IPv4 addresses only (implies 'queryall')
	-6                Connect to IPv6 addresses only (implies 'queryall')
	-t Ns             Query timeout value in seconds (default %v)
	-r N              Maximum # of retries (default %d)
	-printbody        Print body
	-bodyonly         Only print body, no status, headers, etc
	-queryall         Query all server addresses (implies 'noredirect')
	-noredirect       Don't follow redirects
	-sni name         Server Name Indication option
	-header key:val   Send custom request header
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

	if options.ipv6only && options.ipv4only {
		fmt.Printf("ERROR: Cannot specify both -4 and -6. Choose one.\n")
		flag.Usage()
		os.Exit(4)
	}

	if options.ipv6only || options.ipv4only {
		options.queryall = true
	}

	if options.queryall {
		options.noredirect = true
	}

	if *help || (flag.NArg() != 1) {
		if flag.NArg() != 0 {
			fmt.Fprintf(os.Stderr, "ERROR: incorrect number of arguments\n")
		}
		flag.Usage()
		os.Exit(4)
	}

	return flag.Args()[0]
}
