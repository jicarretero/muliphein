package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"fmt"
	"strings"
)

var rq int = 0
var dumpCurl = false

// TargetServer represents a server to forward requests to.
type TargetServer struct {
	URL    string
	Client *http.Client
	isCanisMajor bool
}

// ResponseResult holds the response and timing information for a target server.
type ResponseResult struct {
	URL        string
	StatusCode int
	Body       string
	Duration   time.Duration
	Error      error
}

// Dumps the request received in a CURL statement. In files named /tmp/here-x.req -- 
// It needs the DUMP_AS_CURL variable with the value "yes" in environment.
func DumpCurl(r *http.Request, bodyBytes []byte) {
	if !dumpCurl {
		return
	}

	rq = rq + 1

	s :=  fmt.Sprintf("curl -X %s ${NGSILD_ADDRESS}%s \\\n", r.Method, r.URL.Path)
	for key, value := range r.Header {
		s = fmt.Sprintf("%s -H \"%s: %s\" \\\n",s, key, value[0])
	}
	s = fmt.Sprintf("%s-d '%s'",s, string(bodyBytes))

	log.Printf("\n%s\n", s)

    	tmpFile, err := os.OpenFile( fmt.Sprintf("/tmp/here-%d.req", rq), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer tmpFile.Close()

	// Write the byte array to the file
	if _, err := tmpFile.WriteString(s); err != nil {
		return
	}
}

// Send the received request to any of NGSILD-BROKER or CANIS-MAJOR
func DoSend(target TargetServer, wg *sync.WaitGroup, results *chan ResponseResult, r *http.Request, bodyBytes []byte) {
	defer wg.Done()

	method := r.Method

	// orion-ld has some issues using POST on attributes
	// canis-major does not support PATCH on attibutes
	// ... I cursed and tweaked.
        if r.Method == "POST" &&  strings.HasSuffix(r.URL.Path, "/attrs") && !target.isCanisMajor {
		log.Printf("... changing method to PATCH %s ", target.URL+r.URL.Path)
		method = "PATCH"
	}


	if (method == "GET" || method == "HEAD" || method == "DELETE") && target.isCanisMajor {
		*results <- ResponseResult{
			URL:   target.URL,
			StatusCode: 599,
			Body: "",
		}
		return
	}

	// Start timing the request.
	start := time.Now()

	// Create a new request with a copy of the body.
	req, err := http.NewRequest(method, target.URL+r.URL.Path, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("ERROR [newRequest]%s - %s   ==> %v", method, target.URL+r.URL.Path, err)
		*results <- ResponseResult{
			URL:   target.URL,
			StatusCode: 502,
			Body: err.Error(),
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
		log.Printf("ERROR [target.client.do]%s - %s   ==> %v", method, target.URL+r.URL.Path, err)
		*results <- ResponseResult{
			URL:   target.URL,
			StatusCode: 502,
			Body: err.Error(),
			Error: err,
		}
		return
	}
	defer resp.Body.Close()

	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR [readbody] %s - %s   ==> %v", method, target.URL+r.URL.Path, err)
		*results <- ResponseResult{
			URL:   target.URL,
			StatusCode: 502,
			Body: err.Error(),
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
	log.Println("\n--------------------------------------------------------")
	// Read the original request body.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	DumpCurl(r, bodyBytes)

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
			log.Printf("canisMajorResponse : %s - %d", result.URL, result.StatusCode)
			canisMajorResponse = result
		} else if result.URL == targets[1].URL {
			log.Printf("brokerldResponse : %s - %d", result.URL, result.StatusCode)
			brokerldResponse = result
		}
	}

	// Log the results.
	log.Printf("Response from %s: Status %d, Duration %v",
		canisMajorResponse.URL, canisMajorResponse.StatusCode, canisMajorResponse.Duration)
	log.Printf("Response from %s: Status %d, Duration %v",
		brokerldResponse.URL, brokerldResponse.StatusCode, brokerldResponse.Duration)

	// Decide which response to return to the client.
	// TODO - This is opinionated... need further discussion.
	if canisMajorResponse.StatusCode >= 400 || r.Method == "GET" || r.Method == "HEAD" || r.Method == "DELETE" {
		// If the Canis-Major returns a 40x or 50x error, return its response.
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
	dumpCurl = strings.ToLower(os.Getenv("DUMP_AS_CURL")) == "yes"

	if url_broker == "" || url_canis_major == "" {
		log.Fatalf("Environment variables CANIS_MAJOR_URL and NGSILD_BROKER_URL must be exported")
	}

	log.Printf("Forking between: %s %s", url_canis_major, url_broker)
	// Define the target servers with a 2 seconds timeout each.
	targets := []TargetServer{
		{
			URL:    url_canis_major,
			Client: &http.Client{Timeout: 5 * time.Second},
			isCanisMajor: true,
		},
		{
			URL:    url_broker,
			Client: &http.Client{Timeout: 5 * time.Second},
			isCanisMajor: false,
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
