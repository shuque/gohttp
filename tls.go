package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//
// TLSversion - map TLS verson number to string
//
var TLSversion = map[uint16]string{
	0x0300: "SSL3.0",
	0x0301: "TLS1.0",
	0x0302: "TLS1.1",
	0x0303: "TLS1.2",
	0x0304: "TLS1.3",
}

//
// KeyUsage value to string
//
var KeyUsage = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "DigitalSignature",
	x509.KeyUsageContentCommitment: "ContentCommitment",
	x509.KeyUsageKeyEncipherment:   "KeyEncipherment",
	x509.KeyUsageDataEncipherment:  "DataEncipherment",
	x509.KeyUsageKeyAgreement:      "KeyAgreement",
	x509.KeyUsageCertSign:          "CertSign",
	x509.KeyUsageCRLSign:           "CRLSign",
	x509.KeyUsageEncipherOnly:      "EncipherOnly",
	x509.KeyUsageDecipherOnly:      "DecipherOnly",
}

//
// ExtendedKeyUsage value to string
//
var ExtendedKeyUsage = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                            "Any",
	x509.ExtKeyUsageServerAuth:                     "ServerAuth",
	x509.ExtKeyUsageClientAuth:                     "ClientAuth",
	x509.ExtKeyUsageCodeSigning:                    "CodeSigning",
	x509.ExtKeyUsageEmailProtection:                "EmailProtection",
	x509.ExtKeyUsageIPSECEndSystem:                 "IPSECEndSystem",
	x509.ExtKeyUsageIPSECTunnel:                    "IPSECTunnel",
	x509.ExtKeyUsageIPSECUser:                      "IPSECUser",
	x509.ExtKeyUsageTimeStamping:                   "TimeStamping",
	x509.ExtKeyUsageOCSPSigning:                    "OCSPSigning",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "MicrosoftServerGatedCrypto",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:      "NetscapeServerGatedCrypto",
	x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "MicrosoftCommercialCodeSigning",
	x509.ExtKeyUsageMicrosoftKernelCodeSigning:     "MicrosoftKernelCodeSigning",
}

//
// KU2Strings -
//
func KU2Strings(ku x509.KeyUsage) string {

	var result []string
	for k, v := range KeyUsage {
		if ku&k == k {
			result = append(result, v)
		}
	}
	return strings.Join(result, " ")
}

//
// EKU2Strings -
//
func EKU2Strings(ekulist []x509.ExtKeyUsage) string {

	var result []string
	for _, eku := range ekulist {
		result = append(result, ExtendedKeyUsage[eku])
	}
	return strings.Join(result, " ")
}

//
// KeySizeInBits -
//
func KeySizeInBits(publickey interface{}) int {

	switch v := publickey.(type) {
	case *rsa.PublicKey:
		return v.Size() * 8
	case *ecdsa.PublicKey:
		return v.X.BitLen() + v.Y.BitLen()
	case *ed25519.PublicKey:
		return 256
	default:
		return 0
	}
}

//
// printCertDetails --
// Print some details of the certificate.
//
func printCertDetails(cert *x509.Certificate) {

	fmt.Printf("   X509 version: %d\n", cert.Version)
	fmt.Printf("   Serial#: %x\n", cert.SerialNumber)
	fmt.Printf("   Subject: %v\n", cert.Subject)
	fmt.Printf("   Issuer:  %v\n", cert.Issuer)
	for _, dnsName := range cert.DNSNames {
		fmt.Printf("   SAN dNSName: %s\n", dnsName)
	}
	for _, ipAddress := range cert.IPAddresses {
		fmt.Printf("   SAN IPaddress: %s\n", ipAddress)
	}
	for _, emailAddress := range cert.EmailAddresses {
		fmt.Printf("   SAN emailAddress: %s\n", emailAddress)
	}
	for _, uri := range cert.URIs {
		fmt.Printf("   SAN URI: %v\n", uri)
	}
	fmt.Printf("   Signature Algorithm: %v\n", cert.SignatureAlgorithm)
	fmt.Printf("   PublicKey Algorithm: %v %d-Bits\n",
		cert.PublicKeyAlgorithm, KeySizeInBits(cert.PublicKey))
	fmt.Printf("   Inception:  %v\n", cert.NotBefore)
	fmt.Printf("   Expiration: %v\n", cert.NotAfter)
	fmt.Printf("   KU: %v\n", KU2Strings(cert.KeyUsage))
	fmt.Printf("   EKU: %v\n", EKU2Strings(cert.ExtKeyUsage))
	if cert.BasicConstraintsValid {
		fmt.Printf("   Is CA?: %v\n", cert.IsCA)
	}
	fmt.Printf("   SKI: %x\n", cert.SubjectKeyId)
	fmt.Printf("   AKI: %x\n", cert.AuthorityKeyId)
	fmt.Printf("   OSCP Servers: %v\n", cert.OCSPServer)
	fmt.Printf("   CA Issuer URL: %v\n", cert.IssuingCertificateURL)
	fmt.Printf("   CRL Distribution: %v\n", cert.CRLDistributionPoints)
	fmt.Printf("   Policy OIDs: %v\n", cert.PolicyIdentifiers)
}

//
// printCertChainDetails -
//
func printCertChainDetails(chain []*x509.Certificate) {

	fmt.Printf("## -------------- FULL Certificate Chain ----------------\n")
	for i, cert := range chain {
		fmt.Printf("## Certificate at Depth: %d\n", i)
		printCertDetails(cert)
	}
}

//
// printVerifiedChains -
//
func printVerifiedChains(chains [][]*x509.Certificate) {

	for i, row := range chains {
		fmt.Printf("## Verified Certificate Chain %d:\n", i)
		for j, cert := range row {
			fmt.Printf("  %2d %v\n", j, cert.Subject)
			fmt.Printf("     %v\n", cert.Issuer)
		}
	}
}

func getTLSConfig() *tls.Config {

	tlsconfig := new(tls.Config)

	if options.sni != "" {
		tlsconfig.ServerName = options.sni
	}

	if options.noverify {
		tlsconfig.InsecureSkipVerify = true
	} else if options.cacert != "" {
		cacert, err := ioutil.ReadFile(options.cacert)
		if err != nil {
			log.Fatal(err)
		}
		cacertpool := x509.NewCertPool()
		cacertpool.AppendCertsFromPEM(cacert)
		tlsconfig.RootCAs = cacertpool
	}

	if options.clientcert != "" {
		clientcreds, err := tls.LoadX509KeyPair(options.clientcert, options.clientkey)
		if err != nil {
			log.Fatal(err)
		}
		tlsconfig.Certificates = []tls.Certificate{clientcreds}
	}

	return tlsconfig
}

func printTLSinfo(response *http.Response) {

	if response.TLS == nil {
		fmt.Println("## TLS Connection Info: NONE")
		return
	}
	fmt.Println("## TLS Connection Info:")
	fmt.Printf("   TLS version: %s\n", TLSversion[response.TLS.Version])
	fmt.Printf("   TLS Resumed: %v\n", response.TLS.DidResume)
	fmt.Printf("   TLS CipherSuite: %s\n", tls.CipherSuiteName(response.TLS.CipherSuite))
	fmt.Printf("   TLS ALPN: %s\n", response.TLS.NegotiatedProtocol)
	fmt.Printf("   TLS SNI: %s\n", response.TLS.ServerName)

	if options.showcertchain {
		printCertChainDetails(response.TLS.PeerCertificates)
		printVerifiedChains(response.TLS.VerifiedChains)
	} else if options.showcert {
		fmt.Println("   ## Peer Certificate:")
		printCertDetails(response.TLS.PeerCertificates[0])
	}
}
