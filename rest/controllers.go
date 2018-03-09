package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	k := "REST"
	v := "Home"

	b, err := json.Marshal(map[string]string{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))

	return
}

func Bucket(w http.ResponseWriter, r *http.Request) {
	k := "REST"
	v := "Bucket"

	b, err := json.Marshal(map[string]string{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))

	return
}
