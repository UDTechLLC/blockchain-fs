package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"

	api "bitbucket.org/udt/wizefs/internal/command"
	"github.com/gorilla/mux"
)

func Home(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "HOME")
}

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	var bucketResource BucketResource
	// Decode the incoming Bucket json
	err := json.NewDecoder(r.Body).Decode(&bucketResource)
	if err != nil ||
		bucketResource.Data.Origin == "" {
		displayAppError(w, err, "Invalid Bucket data", http.StatusInternalServerError)
		return
	}

	// Create a Bucket
	if exitCode, err := api.ApiCreate(bucketResource.Data.Origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated,
		&BucketResponse{
			Success: true,
			Message: "Bucket was created!",
			Bucket:  bucketResource,
		})
}

func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// Delete a Bucket
	if exitCode, err := api.ApiDelete(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "Bucket was deleted!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}

func MountBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// FIXME: clean/reset buffer memory?
	var outbuf, errbuf bytes.Buffer

	// Mount a Bucket via mount App
	appPath := projectPath + mountApp
	fmt.Println("appPath:", appPath)
	c := exec.Command(appPath, origin)
	c.Stdout = &outbuf
	c.Stderr = &errbuf

	cerr := c.Start()
	if cerr != nil {
		displayAppError(w, cerr,
			fmt.Sprintf("starting command failed: %v", cerr),
			http.StatusInternalServerError)
		return
	}

	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				displayAppError(w, cerr,
					fmt.Sprintf("wait returned an exit status: %d [%s]", waitstat.ExitStatus(), errbuf.String()[:errbuf.Len()-1]),
					http.StatusInternalServerError)
				return
			}
		} else {
			displayAppError(w, cerr,
				fmt.Sprintf("wait returned an unknown error: %v [%s]", cerr, errbuf.String()[:errbuf.Len()-1]),
				http.StatusInternalServerError)
			return
		}
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "Bucket was mounted!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}

func UnmountBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError)
		return
	}

	// Unmount a Bucket
	if exitCode, err := api.ApiUnmount(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusNoContent)
	respondWithJSON(w, http.StatusOK,
		&BucketResponse{
			Success: true,
			Message: "Bucket was unmounted!",
			Bucket:  BucketResource{Data: BucketModel{Origin: origin}},
		})
}
