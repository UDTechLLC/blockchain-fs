package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"

	"bitbucket.org/udt/wizefs/core/globals"
	core "bitbucket.org/udt/wizefs/core/primitives"
	"github.com/gorilla/mux"
)

var storage *core.Storage

func init() {
	storage = core.NewStorage()
	fmt.Printf("storage buckets: %s\n", storage)
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
		displayAppError(w, err, "Invalid Bucket data",
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	// Create a Bucket
	if exitCode, err := storage.Create(bucketResource.Data.Origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
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
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	// Delete a Bucket
	if exitCode, err := storage.Delete(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
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
			http.StatusInternalServerError, globals.ExitOrigin)
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
			http.StatusInternalServerError, globals.ExitMountPoint)
		return
	}

	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				displayAppError(w, cerr,
					fmt.Sprintf("wait returned an exit status: %d [%s]", waitstat.ExitStatus(), errbuf.String()),
					http.StatusInternalServerError, waitstat.ExitStatus())
				return
			}
		} else {
			displayAppError(w, cerr,
				fmt.Sprintf("wait returned an unknown error: %v [%s]", cerr, errbuf.String()[:errbuf.Len()-1]),
				http.StatusInternalServerError, globals.ExitMountPoint)
			return
		}
	}

	// print mountApp stdout buffer
	//fmt.Println(outbuf.String())

	// FIXME: Mounting the Bucket
	fsinfo, _, err := storage.Config.GetInfoByOrigin(origin)
	bucket, ok := storage.Bucket(origin)
	if err == nil {
		if ok {
			bucket.SetMounted(true)
			bucket.MountPoint = fsinfo.MountpointKey
		}
	}

	//fmt.Printf("Bucket: %+v\n", bucket)

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
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	// Unmount a Bucket
	if exitCode, err := storage.Unmount(origin); err != nil {
		displayAppError(w, err,
			fmt.Sprintf("Error: %s Exit code: %d", err.Error(), exitCode),
			http.StatusInternalServerError, exitCode)
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

func StateBucket(w http.ResponseWriter, r *http.Request) {
	// Get origin from the incoming url
	vars := mux.Vars(r)
	origin := vars["origin"]

	if origin == "" {
		displayAppError(w, nil,
			"Please check request URL!",
			http.StatusInternalServerError, globals.ExitOrigin)
		return
	}

	mounted := false
	bucket, created := storage.Bucket(origin)
	fmt.Printf("Bucket: %+v\n", bucket)
	if created && bucket != nil {
		mounted = bucket.IsMounted()
	}

	respondWithJSON(w, http.StatusOK,
		&BucketStateResponse{
			Success: true,
			Created: created,
			Mounted: mounted,
		})
}
