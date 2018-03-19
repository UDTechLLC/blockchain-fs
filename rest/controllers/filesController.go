package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	//api "bitbucket.org/udt/wizefs/internal/command"
	"github.com/gorilla/mux"
)

func PutFile(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// 32 Mb = 32 << 20
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		displayAppError(w, err,
			"Parsing multipart form was failed! Check your request, please!",
			http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("filename")
	if err != nil {
		displayAppError(w, err,
			"Openning file was failed! Check your request, please!",
			http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := header.Filename
	//fmt.Println("filename:", filename)

	// Copy the file data to the buffer
	var buf bytes.Buffer
	io.Copy(&buf, file)
	defer buf.Reset()

	//if exitCode, err := api.ApiPut(filename, origin, buf.Bytes()); err != nil {
	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError)
		return
	}
	if exitCode, err := bucket.PutFile(filename, buf.Bytes()); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "File " + filename + " was upload to Bucket!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	// Get origin and filename from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]
	filename := vars["filename"]

	if origin == "" || filename == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// Get a File
	//content, exitCode, err := api.ApiGet(filename, origin, "", true)
	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError)
		return
	}
	content, exitCode, err := bucket.GetFile(filename, "", true)
	if err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	if _, err := w.Write(content); err != nil {
		displayAppError(w, err, "", http.StatusInternalServerError)
		return
		//} else {
		//	fmt.Println("Sending", written, "bytes for", filename)
	}
}

func RemoveFile(w http.ResponseWriter, r *http.Request) {
	// Get origin and filename from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]
	filename := vars["filename"]

	if origin == "" || filename == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// Remove a File
	//if exitCode, err := api.ApiRemove(filename, origin); err != nil {
	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError)
		return
	}
	if exitCode, err := bucket.RemoveFile(filename); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "File " + filename + " was removed from Bucket!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}
