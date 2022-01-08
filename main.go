package main

/*
 * Simple HTTP diagnostic tool
 */

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// Version and Program name strings
var Version = "0.0.1"
var progname = path.Base(os.Args[0])

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

// Result structure
type Result struct {
	response     *http.Response
	body         []byte
	responsetime time.Duration
	err          error
}

func printStatus(response *http.Response) {

	fmt.Println("## HTTP Status:")
	fmt.Printf("   HTTP Status: %d %s\n", response.StatusCode, http.StatusText(response.StatusCode))
	fmt.Printf("   HTTP Protocol: %d %d %s\n", response.ProtoMajor, response.ProtoMinor, response.Proto)
	fmt.Printf("   HTTP ContentLength: %d\n", response.ContentLength)
	fmt.Printf("   HTTP Close: %v\n", response.Close)
	fmt.Printf("   HTTP Uncompressed: %v\n", response.Uncompressed)
}

func printHeaders(header http.Header) {

	fmt.Printf("## HTTP Headers:\n")
	for headerkey, headervalue := range header {
		fmt.Printf("   %s: %s\n", headerkey, strings.Join(headervalue, ","))
	}
}

func readResponse(client http.Client, url string) (result *Result) {

	var response *http.Response
	var body []byte
	var err error

	result = new(Result)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		result.err = err
		return
	}
	request.Header.Add("User-Agent", options.useragent)

	if options.username != "" {
		request.SetBasicAuth(options.username, options.password)
	}

	t0 := time.Now()
	response, err = client.Do(request)
	if err != nil {
		result.err = err
		return
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	body, err = ioutil.ReadAll(response.Body)
	result.responsetime = time.Since(t0)
	result.response = response
	result.body = body
	result.err = err
	return
}

func getClient(address string) http.Client {

	client := http.Client{
		Timeout: options.timeout,
	}

	transport := &http.Transport{
		TLSClientConfig:   getTLSConfig(),
		ForceAttemptHTTP2: true,
	}

	if address != "" {
		transport.DialContext = func(ctx context.Context, network, unusedaddress string) (net.Conn, error) {
			dialer := new(net.Dialer)
			dialer.Timeout = options.timeout
			return dialer.Dial(network, address)
		}
	}

	client.Transport = transport

	if options.noredirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func addressString(ipaddress net.IP, port string) string {

	if !strings.Contains(ipaddress.String(), ":") {
		return ipaddress.String() + ":" + port
	}
	return "[" + ipaddress.String() + "]" + ":" + port
}

func url2addressport(urlstring string) (hostname, port string, err error) {

	parsedurl, err := url.Parse(urlstring)
	if err != nil {
		return "", "", err
	}
	hostname = parsedurl.Hostname()
	port = parsedurl.Port()
	if port == "" {
		port = portMap[parsedurl.Scheme]
	}

	return hostname, port, nil
}

func querySingle(urlstring, address string) {

	client := getClient(address)
	result := readResponse(client, urlstring)
	if result.err != nil {
		fmt.Println(result.err)
		return
	}

	if options.printstatus {
		fmt.Printf("## ResponseTime: %v\n", result.responsetime)
		printTLSinfo(result.response)
		printStatus(result.response)
	}

	if options.printheader {
		printHeaders(result.response.Header)
	}

	if options.printbody {
		fmt.Println("## HTTP BODY:")
		fmt.Printf("%s\n", result.body)
	}
}

func getIpList(hostname string) []net.IP {

	iplist, err := net.LookupIP(hostname)
	if err != nil {
		log.Fatal(err)
	}

	if !(options.ipv6only || options.ipv4only) {
		return iplist
	}

	var filteredlist []net.IP

	for _, ipaddress := range iplist {
		isipv4 := (ipaddress.To4() != nil)
		if options.ipv6only && isipv4 {
			continue
		}
		if options.ipv4only && !isipv4 {
			continue
		}
		filteredlist = append(filteredlist, ipaddress)
	}

	return filteredlist
}

func main() {

	urlstring := doFlags()

	hostname, port, err := url2addressport(urlstring)
	if err != nil {
		log.Fatal(err)
	}

	iplist := getIpList(hostname)

	fmt.Printf("URL: %s\nHostname: %s\nPort: %s\n", urlstring, hostname, port)
	fmt.Println("Addresses:")
	for _, ipaddress := range iplist {
		fmt.Printf("\t%s\n", ipaddress)
	}

	if options.queryall {
		for _, ipaddress := range iplist {
			fmt.Printf("\nCONNECT: %s %s ..\n", ipaddress, port)
			querySingle(urlstring, addressString(ipaddress, port))
		}
	} else {
		fmt.Println()
		querySingle(urlstring, "")
	}
}
