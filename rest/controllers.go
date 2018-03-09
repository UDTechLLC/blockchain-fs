package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	api "bitbucket.org/udt/wizefs/internal/command"
)

type BucketModel struct {
	Origin string `json:"origin"`
}

type BucketResource struct {
	Data BucketModel `json:"data"`
}

type appError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	HttpStatus int    `json:"status"`
}

type errorResource struct {
	Data appError `json:"data"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "HOME")
}

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	var bucketResource BucketResource
	// Decode the incoming Bucket json
	err := json.NewDecoder(r.Body).Decode(&bucketResource)
	if err != nil {
		displayAppError(w, err, "Invalid Bucket data", 500)
		return
	}

	log.Println("bucketResource:", bucketResource)

	if exitCode, err := api.ApiCreate(bucketResource.Data.Origin); err != nil {
		displayAppError(w, err, err.Error()+" ExitCode: "+strconv.Itoa(exitCode), 500)
		return
	}

	respondWithJSON(w, http.StatusCreated, "CREATED")

	return
}

func displayAppError(w http.ResponseWriter, handlerError error, message string, code int) {
	errObj := appError{
		Error:      handlerError.Error(),
		Message:    message,
		HttpStatus: code,
	}

	log.Printf("[AppError]: %s\n", handlerError)

	respondWithJSON(w, code, errorResource{Data: errObj})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
