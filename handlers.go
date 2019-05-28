package gemux

import "net/http"

// MethodNotAllowedHandler returns a simple request handler that replies to
// each request with a "405 method not allowed" reply and writes the 405 status
// code.
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	})
}
