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

	fmt.Println("## HTTP Headers:")
	for headerkey, headervalue := range header {
		fmt.Printf("   %s: %s\n", headerkey, strings.Join(headervalue, ","))
	}
	fmt.Println("## End of HTTP Headers.")
}

func readResponse(client http.Client, request *http.Request) (result *Result) {

	var response *http.Response
	var body []byte
	var err error

	result = new(Result)

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

func getRequest(url string) *http.Request {

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("User-Agent", options.useragent)
	for _, header := range options.headers {
		tmp := strings.SplitN(header, ":", 2)
		key := tmp[0]
		val := tmp[1]
		request.Header.Add(key, val)
	}
	return request
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

func querySingle(request *http.Request, address string) {

	client := getClient(address)
	result := readResponse(client, request)
	if result.err != nil {
		fmt.Println(result.err)
		return
	}

	if !options.bodyonly {
		fmt.Printf("## ResponseTime: %v\n", result.responsetime)
		printTLSinfo(result.response)
		printStatus(result.response)
		printHeaders(result.response.Header)
	}

	if options.printbody || options.bodyonly {
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

func prologue(urlstring, hostname, port string, iplist []net.IP) {

	fmt.Printf("URL: %s\nHostname: %s\nPort: %s\n", urlstring, hostname, port)
	fmt.Println("Addresses:")
	for _, ipaddress := range iplist {
		fmt.Printf("\t%s\n", ipaddress)
	}
}

func main() {

	var request *http.Request

	urlstring := doFlags()

	hostname, port, err := url2addressport(urlstring)
	if err != nil {
		log.Fatal(err)
	}
	iplist := getIpList(hostname)

	if !options.bodyonly {
		prologue(urlstring, hostname, port, iplist)
	}

	request = getRequest(urlstring)

	if options.queryall {
		for _, ipaddress := range iplist {
			fmt.Printf("\nCONNECT: %s %s ..\n", ipaddress, port)
			querySingle(request, addressString(ipaddress, port))
		}
	} else {
		fmt.Println()
		querySingle(request, "")
	}
}
