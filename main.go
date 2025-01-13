package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type URL struct {
	ID           string    `json:"id"`
	OriginalUrl  string    `json:"original_url"`
	ShortenedURL string    `json:"shortened_url"`
	CreationDate time.Time `json:"creation_date"`
}

var urlDB = make(map[string]URL)

func generateShortURL(OriginalUrl string) string {

	hasher := md5.New()
	hasher.Write([]byte(OriginalUrl))
	// fmt.Println("hasher result : ", hasher)

	data := hasher.Sum(nil)
	// fmt.Println("hasher result : ", data)

	hash := hex.EncodeToString(data)
	// fmt.Println("hasher result : ", hash[:8])

	return hash[:8]
}

func createURL(originalUrl string) string {
	shortURL := generateShortURL(originalUrl)
	id := shortURL
	urlDB[id] = URL{
		ID:           id,
		OriginalUrl:  originalUrl,
		ShortenedURL: shortURL,
		CreationDate: time.Now(),
	}

	return shortURL
}

func getURL(id string) (URL, error) {
	url, err := urlDB[id]
	if !err {
		return URL{}, errors.New("URL NOT FOUND")
	}
	return url, nil
}

func ShortURLhandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
	}

	fmt.Println("Received URL:", data.URL)

	shortURL := createURL(data.URL)
	// json.NewEncoder(w).Encode(shortURL)

	response := struct {
		ShortenURL string `json:"short_url"`
	}{ShortenURL: shortURL}

	fmt.Println("Shortened URL ID:", response.ShortenURL)
	fmt.Print("\n")
	json.NewEncoder(w).Encode(response)

}

func redirectURLhandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	myurl, err := getURL(id)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, myurl.OriginalUrl, http.StatusFound)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	mux := http.NewServeMux()
    mux.HandleFunc("/", ShortURLhandler)
    mux.HandleFunc("/redirect/", redirectURLhandler)

    handler := enableCORS(mux)

    fmt.Println("Starting server on port 4000")
    err := http.ListenAndServe(":4000", handler)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }

}
