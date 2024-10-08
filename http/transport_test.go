package http_test

import (
	"fmt"
	"net/http"

	"github.com/picatz/simnet"
	simhttp "github.com/picatz/simnet/http"
)

func ExampleClient() {
	client := &http.Client{
		Transport: &simhttp.Transport{
			Dialer: simnet.NewDialer(&simnet.Config{
				// Latency:   100 * time.Millisecond, // Base latency of 100ms
				// Jitter:    50 * time.Millisecond,  // Up to 50ms of additional latency
				// Bandwidth: 256 * 1024,             // Bandwidth limit of 256 KBps
				// LossRate:  0.05,                   // 5% packet loss
				// PartitionedAddrs: map[string]bool{
				// 	"google.com:443": true, // google.com is unreachable
				// },
			}),
		},
	}

	// Use the client to make an HTTP GET request.
	resp, err := client.Get("https://google.com")
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Output the response status and body length.
	fmt.Printf("Response status: %s\n", resp.Status)

	// Output:
	// Response status: 200 OK
}
