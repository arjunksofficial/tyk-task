package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/arjunksofficial/tyk-task/internal/config"
	"github.com/arjunksofficial/tyk-task/internal/middlewares/auth"
	"github.com/arjunksofficial/tyk-task/internal/middlewares/logging"
	"github.com/arjunksofficial/tyk-task/internal/middlewares/ratelimit"
	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	log.Printf("Config loaded: %+v", cfg)
	router := mux.NewRouter()
	router.Use(logging.LoggingMiddleware)
	router.Use(auth.NewAuthMiddleware().AuthMiddleware)
	router.Use(ratelimit.NewRateLimitMiddleware().RateLimitHandler)
	for _, route := range cfg.GetRoutes() {
		target, err := url.Parse(route.Host)
		if err != nil {
			log.Fatalf("Error parsing URL %s: %v", route.Host, err)
		}
		// implement forward proxy for each route
		proxy := httputil.NewSingleHostReverseProxy(target)

		// use regext match for path/* to rouute to the proxy

		router.PathPrefix(route.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Modify the request URL to match the target host
			r.URL.Host = target.Host
			r.URL.Scheme = target.Scheme
			r.Host = target.Host
			r.Header.Set("X-Forwarded-Host", r.Host)
			r.Header.Set("X-Forwarded-For", r.RemoteAddr)
			r.Header.Set("X-Forwarded-Proto", target.Scheme)

			// Log the request for debugging
			log.Printf("Proxying request: %s %s to %s", r.Method, r.URL.Path, target.String())

			// Serve the request using the reverse proxy
			proxy.ServeHTTP(w, r)
		})
		log.Printf("Route registered: %s -> %s", route.Path, route.Host)
	}

	log.Println("Proxy listening on :9000")
	log.Fatal(http.ListenAndServe(":"+cfg.GetPort(), router))
}
