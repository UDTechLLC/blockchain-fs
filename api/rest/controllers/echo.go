package controllers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/disk"
)

var StartTime time.Time

type Ping struct {
	Ping string
}
type Pong struct {
	FreeStorage uint64
	Uptime      int64
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	ping := Ping{}
	pong := Pong{}
	err := json.NewDecoder(r.Body).Decode(&ping)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Bad request")
	}

	if ping.Ping == getOnlyHash("pingstorage") {
		v, _ := disk.Usage("/")
		now := time.Now()
		pong.Uptime = int64(now.Sub(StartTime).Seconds())
		pong.FreeStorage = uint64(v.Free / 1048576)

		respondWithJSON(w, http.StatusOK, pong)
	} else {
		respondWithJSON(w, http.StatusForbidden, "auth required")
	}

}

func getOnlyHash(data string) string {
	h256 := sha256.New()
	out := fmt.Sprintf("%s", data)
	io.WriteString(h256, out)

	return fmt.Sprintf("%x", h256.Sum(nil))
}
