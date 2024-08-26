package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	caCert, err := os.ReadFile(os.Getenv("HTTP_PROBE_CA_CERT"))
	if err != nil {
		log.Fatalf("Unable to read CA certificate: %v", err)
	}

	cert := x509.NewCertPool()
	cert.AppendCertsFromPEM(caCert)

	tlsDialer := &tls.Dialer{
		Config: &tls.Config{RootCAs: cert},
		NetDialer: &net.Dialer{
			KeepAliveConfig: net.KeepAliveConfig{
				Enable:   true,
				Idle:     -1,
				Interval: -1,
			},
		},
	}

	client := &http.Client{
		Transport: &http.Transport{
			// TLSClientConfig: &tls.Config{RootCAs: cert},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return tlsDialer.DialContext(ctx, network, addr)
			},
		},
	}

	internal, err := time.ParseDuration(os.Getenv("HTTP_PROBE_INTERVAL"))
	if err != nil {
		log.Fatalf("Unable to parse HTTP_PROBE_INTERVAL: %v", err)
	}

	t := time.NewTicker(internal)
	for {
		resp, err := client.Get("https://localhost:8443/hostname")
		if err != nil {
			log.Fatalf("Unable to get response: %v", err)
		}

		fmt.Println("Response", resp.StatusCode)
		for k, v := range resp.Header {
			fmt.Printf("%v: %v\n", k, v)
		}

		fmt.Println("Wait for next probe in", internal)
		fmt.Println()

		<-t.C
	}
}
