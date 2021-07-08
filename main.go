package main

import (
	"fmt"
	"net/http"

	"github.com/nickolasoverlay/aniview_encoder/v2/src"
)

func main() {
	src.Init()

	fmt.Println("AniviewEncoder")
	fmt.Println("-> OUTPUT_PATH was set to", src.GetEncoderEnv().OutputPath)

	http.HandleFunc("/schedule", src.Schedule)
	http.HandleFunc("/stats", src.QueueStats)
	http.ListenAndServe(":6000", nil)
}
