package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	api "bitbucket.org/udt/wizefs/internal/command"
	"github.com/gorilla/mux"
)

const (
	packagePath = "rest"
	mountApp    = "cmd/wizefs_mount/wizefs_mount"
)

var (
	projectPath = getProjectPath()
)

func getProjectPath() string {
	_, testFilename, _, _ := runtime.Caller(0)
	idx := strings.Index(testFilename, packagePath)
	return testFilename[0:idx]
}

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
	if err != nil ||
		bucketResource.Data.Origin == "" {

		displayAppError(w, err, "Invalid Bucket data", 500)
		return
	}

	// Create a Bucket
	if exitCode, err := api.ApiCreate(bucketResource.Data.Origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			500)
		return
	}

	respondWithJSON(w, http.StatusCreated, "CREATED")
}

func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]
	// Delete a Bucket
	if exitCode, err := api.ApiDelete(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func MountBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	// Mount a Bucket via mount App
	appPath := projectPath + mountApp
	c := exec.Command(appPath, origin)
	cerr := c.Start()
	if cerr != nil {
		displayAppError(w, cerr,
			fmt.Sprintf("starting command failed: %v", cerr),
			500)
		return
	}

	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				displayAppError(w, cerr,
					fmt.Sprintf("wait returned an exit status: %d", waitstat.ExitStatus()),
					500)
			}
		} else {
			displayAppError(w, cerr,
				fmt.Sprintf("wait returned an unknown error: %v", cerr),
				500)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func UnmountBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]
	// Unmount a Bucket
	if exitCode, err := api.ApiUnmount(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	w.Write([]byte("\n"))
}
