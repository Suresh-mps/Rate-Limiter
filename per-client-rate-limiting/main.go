package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

// perClientRateLimiter is a middleware that applies rate limiting per client IP.
func perClientRateLimiter(next func(writer http.ResponseWriter, request *http.Request)) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client) // Map to store rate limiters for each client IP.
	)

	// Background goroutine to clean up stale clients periodically.
	go func() {
		for {
			time.Sleep(time.Minute) // Run cleanup every minute.
			mu.Lock()
			for ip, client := range clients {
				// Remove clients that haven't been seen for more than 3 minutes.
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Lock the mutex to safely access the clients map.
		mu.Lock()
		if _, found := clients[ip]; !found {
			// Create a new rate limiter for the client if not already present.
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		clients[ip].lastSeen = time.Now() // Update the last seen timestamp.
		if !clients[ip].limiter.Allow() {
			mu.Unlock()

			// Respond with a rate limit exceeded message.
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()

		// Call the next handler if the request is allowed.
		next(w, r)
	})
}

// endpointHandler is the main handler for the /ping endpoint.
func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Success",
		Body:   "Hi there, this is a rate-limited endpoint.",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	// Attach the rate limiter middleware to the /ping endpoint.
	http.Handle("/ping", perClientRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Println("There was an error listening on port :8000", err)
	}
}