// Determines which websites can call FBBridge Proxy using browser JS.
package middleware

import "net/http"

type CORS struct {
	next           http.Handler
	allowedOrigins map[string]struct{}
}

func NewCORS(next http.Handler, allowedOrigins []string) *CORS {
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origins[origin] = struct{}{}
	}

	return &CORS{
		next:           next,
		allowedOrigins: origins,
	}
}

func (c *CORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/events" {
		c.next.ServeHTTP(w, r)
		return
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		c.next.ServeHTTP(w, r)
		return
	}

	if _, allowed := c.allowedOrigins[origin]; !allowed {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Add("Vary", "Origin")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	c.next.ServeHTTP(w, r)
}
