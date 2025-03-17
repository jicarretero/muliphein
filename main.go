package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// TargetServer represents a server to forward requests to.
type TargetServer struct {
	URL    string
	Client *http.Client
}

// ResponseResult holds the response and timing information for a target server.
type ResponseResult struct {
	URL        string
	StatusCode int
	Body       string
	Duration   time.Duration
	Error      error
}

func DoSend(target TargetServer, wg *sync.WaitGroup, results *chan ResponseResult, r *http.Request, bodyBytes []byte) {
	defer wg.Done()

	// Start timing the request.
	start := time.Now()

	// Create a new request with a copy of the body.
	req, err := http.NewRequest(r.Method, target.URL+r.URL.Path, bytes.NewReader(bodyBytes))
	if err != nil {
		*results <- ResponseResult{
			URL:   target.URL,
			Error: err,
		}
		return
	}

	// Copy headers from the original request.
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send the request to the target server.
	resp, err := target.Client.Do(req)
	if err != nil {
		*results <- ResponseResult{
			URL:   target.URL,
			Error: err,
		}
		return
	}
	defer resp.Body.Close()

	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		*results <- ResponseResult{
			URL:   target.URL,
			Error: err,
		}
		return
	}

	// Calculate the duration.
	duration := time.Since(start)

	// Send the result to the channel.
	*results <- ResponseResult{
		URL:        target.URL,
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Duration:   duration,
	}
}

// ForwardRequest forwards the incoming request to multiple target servers.
func ForwardRequest(targets []TargetServer, w http.ResponseWriter, r *http.Request) {
	// Read the original request body.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	log.Printf("%s  %s", r.Method, r.URL)
	defer r.Body.Close()

	// Create a wait group to wait for all requests to complete.
	var wg sync.WaitGroup
	wg.Add(len(targets))

	// Channel to collect results from each target.
	results := make(chan ResponseResult, len(targets))

	// Forward the request to each target server concurrently.
	for _, target := range targets {
		go DoSend(target, &wg, &results, r, bodyBytes)
	}

	// Wait for all requests to complete.
	wg.Wait()
	close(results)

	// Collect the results.
	var (
		canisMajorResponse ResponseResult
		brokerldResponse   ResponseResult
	)

	for result := range results {
		if result.URL == targets[0].URL {
			canisMajorResponse = result
		} else if result.URL == targets[1].URL {
			brokerldResponse = result
		}
	}

	// Log the results.
	log.Printf("Response from %s: Status %d, Duration %v, Body: %s",
		canisMajorResponse.URL, canisMajorResponse.StatusCode, canisMajorResponse.Duration, canisMajorResponse.Body)
	log.Printf("Response from %s: Status %d, Duration %v, Body: %s",
		brokerldResponse.URL, brokerldResponse.StatusCode, brokerldResponse.Duration, brokerldResponse.Body)

	// Decide which response to return to the client.
	if brokerldResponse.StatusCode >= 400 || r.Method == "GET" || r.Method == "POST" {
		// If the broker-ngsild returns a 40x or 50x error, return its response.
		w.WriteHeader(brokerldResponse.StatusCode)
		w.Write([]byte(brokerldResponse.Body))
	} else {
		// Otherwise, return the response of Canis Major
		w.WriteHeader(canisMajorResponse.StatusCode)
		w.Write([]byte(canisMajorResponse.Body))
	}
}

func main() {
	url_canis_major := os.Getenv("CANIS_MAJOR_URL")
	url_broker := os.Getenv("NGSILD_BROKER_URL")

	if url_broker == "" || url_canis_major == "" {
		log.Fatalf("Environment variables CANIS_MAJOR_URL and NGSILD_BROKER_URL must be exported")
	}

	log.Printf("Forking between: %s %s", url_canis_major, url_broker)
	// Define the target servers with a 2 seconds timeout each.
	targets := []TargetServer{
		{
			URL:    url_canis_major,
			Client: &http.Client{Timeout: 2 * time.Second},
		},
		{
			URL:    url_broker,
			Client: &http.Client{Timeout: 2 * time.Second},
		},
	}

	// Set up the HTTP server.
	http.HandleFunc("/", func(writer http.ResponseWriter, reader *http.Request) {
		// Forward the request to all target servers.
		ForwardRequest(targets, writer, reader)
	})

	// Start the server.
	log.Println("Starting proxy server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}
