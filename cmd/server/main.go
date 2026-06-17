package main

import (
	"fmt"
	"net/http"

	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/httpapi"
	"stream-platform/internal/live"
	"stream-platform/internal/storage"
	"stream-platform/internal/vod"
)

func main() {
	runner := ffmpeg.NewRunner()
	store := storage.NewStore("data")

	liveManager := live.NewManager(runner, store)
	liveService := live.NewService(liveManager)
	vodService := vod.NewService(store)

	server := httpapi.NewServer(liveService, vodService, store)

	fmt.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		panic(err)
	}
}
