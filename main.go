package main

import (
	"net/http"

	"github.com/nickolasoverlay/aniview_encoder/v2/src"
)

func main() {
	src.Init()

	http.HandleFunc("/schedule", src.Schedule)
	http.HandleFunc("/stats", src.QueueStats)
	http.ListenAndServe(":6000", nil)
}
