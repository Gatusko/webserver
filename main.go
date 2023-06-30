package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()
	//mux.Handle("/", http.FileServer(http.Dir(".")))
	r.Use(middlewareCors)
	r.Mount("/api", apiRouter(&newApiConfig))
	//httpFileHandle := http.StripPrefix("/app", http.FileServer(http.Dir("./app")))
	assetsHandler := http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./app/assets")))
	r.Handle("/app*", newApiConfig.middleWareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app")))))
	r.Handle("/app/assets/*", assetsHandler)
	server := http.Server{}
	server.Addr = "localhost:8080"
	server.Handler = r
	server.ListenAndServe()
}

func apiRouter(config *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", healthHandler)
	r.Get("/metrics", config.getCount)
	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("CORS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		}
		next.ServeHTTP(w, r)
	})
}
