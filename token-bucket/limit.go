package main
import (
	"encoding/json"
	"net/http"

	"golang.org/x/time/rate"
)

// tokenBucketRateLimiter is a middleware function that implements a token bucket
// rate-limiting algorithm to control the rate of incoming HTTP requests.
// 
// It uses the "rate.NewLimiter" from the "golang.org/x/time/rate" package to
// enforce the rate limit. The limiter allows a maximum of 20 requests per second
// with a burst capacity of 40 requests.
//
// If the rate limit is exceeded, the middleware responds with an HTTP 429
// (Too Many Requests) status code and a JSON message indicating that the API
// is at capacity. Otherwise, it forwards the request to the next handler.
//
// Parameters:
// - next: The next HTTP handler to be called if the request is allowed.
//
// Returns:
// - An http.Handler that enforces the rate limit.
func tokenBucketRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(20, 40)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		} else {
			next(w, r)
		}
	})
}