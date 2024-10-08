package httptest_test

import (
	"fmt"
	"net/http"

	"github.com/picatz/simnet"
	"github.com/picatz/simnet/http/httptest"
)

func ExampleServer() {
	// Define the network conditions.
	cfg := &simnet.Config{
		// Latency:   100 * time.Millisecond,
		// Jitter:    50 * time.Millisecond,
		// Bandwidth: 256 * 1024, // 256 KBps
		LossRate: 0.05, // 5% packet loss
	}

	// Create a handler for the server.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, simnet!")
	})

	// Start a new simulated HTTP server.
	server := httptest.NewServer(cfg, handler)
	defer server.Close()

	// Create a client to make requests to the server.
	client := server.Client()

	// Make a request to the server.
	resp, err := client.Get(server.URL())
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)

	// Output:
	// Response status: 200 OK
}
