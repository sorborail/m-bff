package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
	"github.com/sorborail/m-bff/bff"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	hsAddr := flag.String("hsaddr", "localhost:50051", "The gRPC address for highscore service")
	geAddr := flag.String("geaddr", "localhost:60051", "The gRPC address for gameengine service")
	srvAddr := flag.String("srvaddr", "localhost:8081", "HTTP server address")
	flag.Parse()

	connHs, err := grpc.Dial(*hsAddr, grpc.WithInsecure())
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to deal highscore server: %v", err)
	}
	hsCl, err := bff.NewGameClient(connHs)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Error to create highscore client service")
	}

	connGe, err := grpc.Dial(*geAddr, grpc.WithInsecure())
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to deal gameengine server: %v", err)
	}
	geCl, err := bff.NewGameEngineClient(connGe)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Error to create gameengine client service")
	}
	gr := bff.NewGameResource(hsCl, geCl)

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.GET("/geths", gr.GetHighScore)
	router.GET("/seths/:hs", gr.SetHighScore)
	router.GET("/getsize", gr.GetSize)
	router.GET("/setscore/:score", gr.SetScore)

	srv := &http.Server{
		Addr:    *srvAddr,
		Handler: router,
	}

	go func() {
		zlog.Info().Msg("HTTP Server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msgf("Error start listen on %v", *srvAddr)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//Block until a signal is received
	<- ch
	zlog.Info().Msg("Shutdown HTTP Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zlog.Fatal().Err(err).Msg("Error while shutting down http server")
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		zlog.Info().Msg("timeout of 5 seconds")
		err := connHs.Close()
		if err != nil {
			zlog.Error().Err(err).Str("highscore address", *hsAddr).Msg("Failed to close connection")
		} else {
			zlog.Info().Str("highscore address", *hsAddr).Msg("Connection closed")
		}
		err = connGe.Close()
		if err != nil {
			zlog.Error().Err(err).Str("gameengine address", *geAddr).Msg("Failed to close connection")
		} else {
			zlog.Info().Str("gameengine address", *geAddr).Msg("Connection closed")
		}
	}
	zlog.Info().Msg("HTTP Server exiting.")
}
