package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v7"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

// endpointHandler handles incoming requests and sends a JSON response.
func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hi there, this is a tollbooth rate-limited endpoint.",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	// Define the rate limit exceeded message.
	message := Message{
		Status: "Request Failed",
		Body:   "The API is at capacity, try again later.",
	}
	jsonMessage, _ := json.Marshal(message)

	// Create a new rate limiter allowing 1 request per second.
	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	tlbthLimiter.SetMessageContentType("application/json")
	tlbthLimiter.SetMessage(string(jsonMessage))

	// Attach the rate limiter to the /ping endpoint.
	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Println("There was an error listening on port :8000", err)
	}
}