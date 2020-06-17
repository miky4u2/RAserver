package webserver

import (
	"github.com/miky4u2/RAserver/server/config"
	"github.com/miky4u2/RAserver/server/webserver/handler"
	"github.com/miky4u2/RAserver/server/webserver/handler/api"
	"golang.org/x/time/rate"
	"net/http"
	"path/filepath"
)

// Start HTTP Server
//
func Start() error {

	// Set routing
	mux := http.NewServeMux()
	mux.Handle("/favicon.ico", http.NotFoundHandler()) // returns a 404
	mux.HandleFunc("/", handler.Index)
	mux.HandleFunc("/agent/update", handler.AgentUpdate)
	mux.HandleFunc("/agent/ctl", handler.AgentCtlHandler)
	mux.HandleFunc("/server/ctl", handler.ServerCtlHandler)
	mux.HandleFunc("/api/download", api.Download)

	// TLS certificate and key paths
	cert := filepath.Join(config.AppBasePath, `conf`, `cert.pem`)
	key := filepath.Join(config.AppBasePath, `conf`, `key.pem`)

	// Launch TLS HTTP server
	err := http.ListenAndServeTLS(config.Settings.BindIP+`:`+config.Settings.BindPort, cert, key, limit(mux))
	if err != nil {
		return err
	}
	return err
}

// Middleware ratel limiter 5 request per second with burst of 30
//
var limiter = rate.NewLimiter(5, 30)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
