package main

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// хост
const host string = ""

// порт
const port string = "8080"

var pairs = make(map[string]string)

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/favicon.ico", icoHandler)

	fmt.Println("Service listening on", host, ":", port)

	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}

func icoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		id := r.URL.Path[1:]

		url, ok := getUrl(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("Error, no such link"))
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case http.MethodPost:
		body, _ := ioutil.ReadAll(r.Body)
		id := string(body)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if len(id) > 2048 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error, link cannot be longer than 2048 characters"))
			return
		}

		url, ok := getUrl(id)
		var err error
		if !ok {
			url, err = shortener(id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error, failed to generate shortened URL"))
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(url))

	default:
		http.Error(w, "Error, only GET and POST methods can be used", http.StatusMethodNotAllowed)

	}
}

func getUrl(id string) (string, bool) {
	if len(id) <= 0 {
		return "", false
	}
	url, ok := pairs[id]
	if !ok {
		return "", false
	}

	return url, true
}

func shortener(s string) (string, error) {
	hasher := crypto.MD5.New()
	if _, err := hasher.Write([]byte(s)); err != nil {
		return "", fmt.Errorf("Reduction error URL: %v", err)
	}
	hash := string(hasher.Sum([]byte{}))
	hash = hash[len(hash)-5:]
	id := base64.StdEncoding.EncodeToString([]byte(hash))
	id = strings.ToLower(id)
	id = strings.ReplaceAll(id, "=", "")
	id = strings.ReplaceAll(id, "/", "_")

	pairs[id] = s

	return id, nil
}
