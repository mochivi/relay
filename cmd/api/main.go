package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
)

func mustHostname() string {
	h, _ := os.Hostname()
	if h == "" {
		return "unknown"
	}
	return h
}

// API for testing proxy
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	route := os.Getenv("ROUTE")
	if route == "" {
		route = "/"
	}
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "api"
	}

	// Ensure route prefix for matching (e.g. /admin matches /admin and /admin/...)
	route = strings.TrimSuffix(route, "/")
	if route == "" {
		route = "/"
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		randint_64, _ := rand.Int(rand.Reader, big.NewInt(1000))
		time.Sleep(time.Duration(int(randint_64.Int64())) * time.Millisecond)

		path := r.URL.Path
		// Route "/" means catch-all: serve every path
		if route != "/" {
			if path != route && !strings.HasPrefix(path, route+"/") {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		body := map[string]string{
			"ok":      "true",
			"service": serviceName,
			"route":   route,
			"path":    path,
			"host":    mustHostname(), // identifies instance when multiple backends use same port (e.g. api1:3001, api2:3001)
		}
		_ = json.NewEncoder(w).Encode(body)
	})

	addr := "0.0.0.0:" + port
	log.Printf("Listening on %s, serving route %s (service=%s)", addr, route, serviceName)
	log.Fatal(http.ListenAndServe(addr, handler))
}
