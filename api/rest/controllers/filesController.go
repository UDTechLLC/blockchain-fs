package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"bitbucket.org/udt/wizefs/core/globals"
)

func PutFile(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	// 32 Mb = 32 << 20
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		displayAppError(w, err,
			"Parsing multipart form was failed! Check your request, please!",
			http.StatusInternalServerError, globals.ExitFile)
		return
	}

	file, header, err := r.FormFile("filename")
	if err != nil {
		displayAppError(w, err,
			"Openning file was failed! Check your request, please!",
			http.StatusInternalServerError, globals.ExitFile)
		return
	}
	defer file.Close()

	filename := header.Filename
	//fmt.Println("filename:", filename)

	// Copy the file data to the buffer
	var buf bytes.Buffer
	io.Copy(&buf, file)
	defer buf.Reset()

	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}
	if exitCode, err := bucket.PutFile(filename, buf.Bytes()); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
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

func Put(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	var putResource PutResource
	// Decode the incoming Put json
	err := json.NewDecoder(r.Body).Decode(&putResource)
	if err != nil ||
		putResource.Data.Filename == "" {
		displayAppError(w, err, "Invalid Put data",
			http.StatusInternalServerError, globals.ExitFile)
		return
	}

	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}
	if exitCode, err := bucket.PutFile(putResource.Data.Filename, []byte(putResource.Data.Content)); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
		return
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "File " + putResource.Data.Filename + " was upload to Bucket!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	// Get origin and filename from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]
	filename := vars["filename"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	if filename == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError, globals.ExitFile)
		return
	}

	// Get a File
	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}
	content, exitCode, err := bucket.GetFile(filename, "", true)
	if err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	if _, err := w.Write(content); err != nil {
		displayAppError(w, err, "", http.StatusInternalServerError, globals.ExitFile)
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
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	// Remove a File
	bucket, ok := storage.Bucket(origin)
	if !ok {
		displayAppError(w, nil,
			fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin),
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}
	if exitCode, err := bucket.RemoveFile(filename); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
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
