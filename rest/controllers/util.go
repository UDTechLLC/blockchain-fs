package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

const (
	packagePath = "rest"
	mountApp    = "cmd/wizefs_mount/wizefs_mount"
)

var (
	projectPath = getProjectPath()
)

type BucketModel struct {
	Origin string `json:"origin"`
}

type BucketResource struct {
	Data BucketModel `json:"data"`
}

type BucketResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Bucket  BucketResource `json:"bucket"`
}

type PutModel struct {
	Filename string `json:"name"`
	Content  string `json:"content"`
}

type PutResource struct {
	Data PutModel `json:"data"`
}

type appError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	HttpStatus int    `json:"status"`
}

type errorResource struct {
	Data appError `json:"data"`
}

func getProjectPath() string {
	_, testFilename, _, _ := runtime.Caller(0)
	idx := strings.Index(testFilename, packagePath)
	return testFilename[0:idx]
}

func displayAppError(w http.ResponseWriter, handlerError error, message string, code int) {
	errObj := appError{
		Error:      handlerError.Error(),
		Message:    message,
		HttpStatus: code,
	}

	fmt.Printf("[app error]: %s\n", handlerError)

	respondWithJSON(w, code, errorResource{Data: errObj})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
