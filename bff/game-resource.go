package bff

import (
	"context"
	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
	gameenginepb "github.com/sorborail/m-apis/game-enginepb/v1"
	highscorepb "github.com/sorborail/m-apis/highscorepb/v1"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
	"time"
)

type gameResource struct {
	gameClient highscorepb.GameClient
	gameEngineClient gameenginepb.GameEngineClient
}

func NewGameResource(gmcl highscorepb.GameClient, gecl gameenginepb.GameEngineClient) *gameResource {
	return &gameResource{
		gameClient:       gmcl,
		gameEngineClient: gecl,
	}
}

func NewGameClient(conn *grpc.ClientConn) (highscorepb.GameClient, error) {
	cl := highscorepb.NewGameClient(conn)
	if cl == nil {
		zlog.Error().Msg("Error creating instance of Highscore client")
		return nil, nil
	} else {
		zlog.Info().Msg("Highscore client is created")
		return cl, nil
	}
}

func NewGameEngineClient(conn *grpc.ClientConn) (gameenginepb.GameEngineClient, error) {
	cl := gameenginepb.NewGameEngineClient(conn)
	if cl == nil {
		zlog.Error().Msg("Error in creating instance of Gameengine client service")
		return nil, nil
	} else {
		zlog.Info().Msg("Gameengine client is created")
		return cl, nil
	}
}

func (gr *gameResource) SetHighScore(ct *gin.Context) {
	hsparam := ct.Param("hs")
	hs, err := strconv.ParseFloat(hsparam, 64)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to convert hs param to float from http request")
		return
	}
	zlog.Info().Msg("Begin SetHighscore request...")
	req := &highscorepb.SetHighScoreRequest{HighScore: hs}
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	res, err := gr.gameClient.SetHighScore(ctx, req)
	if err != nil {
		zlog.Error().Err(err).Msg("error happened while SetHighscore response")
	}
	zlog.Info().Interface("SetHighscore", res.GetStatus()).Msg("Value from SetHighscore service")
	ct.JSONP(200, gin.H{
		"hs": hs,
	})
}

func (gr *gameResource) GetHighScore(ct *gin.Context) {
	zlog.Info().Msg("Begin GetHighscore request...")
	req := &highscorepb.GetHighScoreRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	res, err := gr.gameClient.GetHighScore(ctx, req)
	if err != nil {
		zlog.Error().Err(err).Msg("error happened while GetHighscore response")
		return
	}
	zlog.Info().Interface("GetHighscore", res.GetHighScore()).Msg("Value from GetHighscore service")
	hs := strconv.FormatFloat(res.GetHighScore(), 'e', -1, 64)
	ct.JSONP(200, gin.H{
		"hs": hs,
	})
}

func (gr *gameResource) GetSize(ct *gin.Context) {
	zlog.Info().Msg("Begin GetSize request...")
	req := &gameenginepb.GetSizeRequest{}
	//ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	//defer cancel()
	res, err := gr.gameEngineClient.GetSize(context.Background(), req)
	if err != nil {
		zlog.Error().Err(err).Msg("error happened while GetSize response")
		ct.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	zlog.Info().Interface("GetSize", res.GetSize()).Msg("Value from GetSize service")
	sz := strconv.FormatFloat(res.GetSize(), 'e', -1, 64)
	ct.JSONP(http.StatusOK, gin.H{
		"size": sz,
	})
}

func (gr *gameResource) SetScore(ct *gin.Context) {
	zlog.Info().Msg("Begin SetScore request...")
	scparam := ct.Param("score")
	sc, err := strconv.ParseFloat(scparam, 64)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to convert sc param to float from http request")
		return
	}
	req := &gameenginepb.SetScoreRequest{Score: sc}
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	res, err := gr.gameEngineClient.SetScore(ctx, req)
	if err != nil {
		zlog.Error().Err(err).Msg("error happened while SetScore response")
		return
	}
	zlog.Info().Interface("SetScore", res.GetResult()).Msg("Value from SetScore service")
	ct.JSONP(200, gin.H{
		"score": sc,
	})
}
