package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"stream-platform/internal/channel"
	"stream-platform/internal/config"
	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/httpapi"
	"stream-platform/internal/live"
	"stream-platform/internal/storage"
	"stream-platform/internal/vod"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		panic(err)
	}

	runner := ffmpeg.NewRunner()
	store := storage.NewStore(cfg.DataDir)

	liveRepo := live.NewPostgresRepository(db)

	liveManager := live.NewManager(runner, store, liveRepo)
	liveService := live.NewService(liveManager)
	vodService := vod.NewService(store)

	channelRepo := channel.NewPostgresRepository(db)
	channelService := channel.NewService(channelRepo)

	server := httpapi.NewServer(liveService, vodService, channelService, store)

	fmt.Println("server listening on", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		panic(err)
	}
}
