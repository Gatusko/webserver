package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

// The middle wares need to make a new function of server next
func (cfg *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits = cfg.fileserverHits + 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
}

func main() {
	newApiConfig := apiConfig{}
	mux := http.NewServeMux()

	//mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/healthz", healthHandler)
	corsMux := middlewareCors(mux)
	//httpFileHandle := http.StripPrefix("/app", http.FileServer(http.Dir("./app")))
	mux.Handle("/app/", newApiConfig.middleWareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app")))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./app/assets"))))
	mux.HandleFunc("/metrics", newApiConfig.getCount)
	server := http.Server{}
	server.Addr = "localhost:8080"
	server.Handler = corsMux
	server.ListenAndServe()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		}
		next.ServeHTTP(w, r)
	})
}
