package main

import (
	"encoding/json"
	"fmt"
	"github.com/Gatusko/webserver/internal"
	"github.com/Gatusko/webserver/structs"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}

type response struct {
	Error       string `json:"error,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
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

func (cfg *apiConfig) getAdminCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("\t<html>\n\n\t<body>\n\t<h1>Welcome, Chirpy Admin</h1>\n\t<p>Chirpy has been visited %d times!</p>\n\t</body>\n\n\t</html>", cfg.fileserverHits)))
}

var Database *internal.DB

func main() {
	memory, err := internal.NewDB(".")
	if err != nil {
		log.Fatalf("Issue loading the databse : %s", err)
		return
	}
	Database = memory
	newApiConfig := apiConfig{}
	r := chi.NewRouter()
	//mux.Handle("/", http.FileServer(http.Dir(".")))
	r.Use(middlewareCors)
	r.Mount("/api", apiRouter(&newApiConfig))
	r.Mount("/admin", adminRouter(&newApiConfig))
	//httpFileHandle := http.StripPrefix("/app", http.FileServer(http.Dir("./app")))
	assetsHandler := newApiConfig.middleWareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	r.Handle("/app", assetsHandler)
	r.Handle("/app/*", assetsHandler)
	server := http.Server{}
	server.Addr = "localhost:8080"
	server.Handler = r
	server.ListenAndServe()
}

func apiRouter(config *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", healthHandler)
	r.Post("/chirps", validateHandler)
	r.Get("/chirps", getAllChirps)
	r.Get("/chirps/{chirpId}", getChirp)
	r.Post("/users", createUser)
	//r.Get("/metrics", config.getCount)
	return r
}

func createUser(w http.ResponseWriter, r *http.Request) {
	user := structs.User{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	user, err = Database.NewUser(user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
	return
}

func getChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := chi.URLParam(r, "chirpId")
	chirpInt, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	myChirp, err := Database.GetChirp(chirpInt)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("%s", err))
		return
	}
	respondWithJSON(w, http.StatusOK, myChirp)
	return
}

func getAllChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := Database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting all chirps :  %s", err))
		return
	}
	respondWithJSON(w, http.StatusOK, allChirps)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	chir := structs.Chirpy{}
	err := decoder.Decode(&chir)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode parameters")
		return
	}
	if len(chir.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	const badWordsK = "kerfuffle"
	const badWordsS = "sharbert"
	const badWordsF = "fornax"
	spplitedString := strings.Split(chir.Body, " ")
	for i, split := range spplitedString {
		split = strings.ToLower(split)
		if split == badWordsS || split == badWordsK || split == badWordsF {
			spplitedString[i] = "****"
		}
	}
	chir, err = Database.CreateChirp(chir.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error : %s", err))
		return
	}
	respondWithJSON(w, http.StatusCreated, chir)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5xx error: %s", msg)
	}
	resp := response{}
	resp.Error = msg
	respondWithJSON(w, code, resp)
}

func adminRouter(config *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/metrics", config.getAdminCount)
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
