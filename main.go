package main

/*
docs: https://pkg.go.dev/net/http
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

// Version and Program name strings
var Version = "0.0.1"
var progname = path.Base(os.Args[0])

func printStatus(response *http.Response) {

	fmt.Println("## Status:")
	fmt.Printf("   HTTP Status: %d\n", response.StatusCode)
	fmt.Printf("   ProtoMajor: %d\n", response.ProtoMajor)
	fmt.Printf("   ProtoMinor: %d\n", response.ProtoMinor)
	fmt.Printf("   ContentLength: %d\n", response.ContentLength)
	fmt.Printf("   Close: %v\n", response.Close)
	fmt.Printf("   Uncompressed: %v\n", response.Uncompressed)
}

func printHeaders(header http.Header) {

	fmt.Printf("## Headers:\n")
	for headerkey, headervalue := range header {
		fmt.Printf("   %s: %s\n", headerkey, strings.Join(headervalue, ","))
	}
}

func readResponse(client http.Client, url string) (response *http.Response, body []byte, err error) {

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("User-Agent", options.useragent)
	response, err = client.Do(request)
	if err != nil {
		return
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	body, err = ioutil.ReadAll(response.Body)
	return
}

func getClient() http.Client {

	client := http.Client{
		Timeout: options.timeout,
		Transport: &http.Transport{
			TLSClientConfig: getTLSConfig(),
		},
	}

	if options.noredirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func main() {

	url := doFlags()

	client := getClient()
	response, body, err := readResponse(client, url)
	if err != nil {
		log.Fatal(err)
	}

	if options.printstatus {
		printStatus(response)
		printTLSinfo(response)
	}

	if options.printheader {
		printHeaders(response.Header)
	}

	if options.printbody {
		fmt.Printf("%s\n", body)
	}
}
