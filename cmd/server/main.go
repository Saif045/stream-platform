package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stream-platform/internal/auth"
	"stream-platform/internal/channel"
	"stream-platform/internal/config"
	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/httpapi"
	"stream-platform/internal/live"
	"stream-platform/internal/storage"
	"stream-platform/internal/user"
	"stream-platform/internal/vod"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	auth.SetSecret(cfg.JWTSecret)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	ffmpegRunner := ffmpeg.NewRunner()
	store := storage.NewStore(cfg.DataDir)

	liveStore := live.NewPostgresStore(db)
	liveManager := live.NewManager(ffmpegRunner, store, liveStore)

	vodService := vod.NewService(store)

	channelStore := channel.NewPostgresStore(db)
	channelService := channel.NewService(channelStore)

	liveService := live.NewService(liveStore, liveManager, channelService)

	userStore := user.NewPostgresStore(db)
	userService := user.NewService(userStore)

	server := httpapi.NewServer(liveService, vodService, channelService, userService, store, cfg.HookSecret)

	fmt.Println("server listening on", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		panic(err)
	}
}
