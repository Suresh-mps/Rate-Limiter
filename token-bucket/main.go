package main

import (
	"encoding/json"
	"log"
	"net/http"
)


type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

// endpointHandler handles incoming HTTP requests and sends a JSON response.
func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json") // Set response content type to JSON.
	writer.WriteHeader(http.StatusOK)                      // Set HTTP status code to 200 OK.
	message := Message{
		Status: "Successful",
		Body:   "Hi there, this is a rate-limited endpoint.",
	}
	err := json.NewEncoder(writer).Encode(&message) // Encode the message as JSON and write it to the response.
	if err != nil {
		return
	}
}

func main() {
	// Attach the rate limiter middleware to the /ping endpoint.
	http.Handle("/ping", tokenBucketRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8000", nil) // Start the HTTP server on port 8000.
	if err != nil {
		log.Println("There was an error listening on port :8000", err)
	}
}