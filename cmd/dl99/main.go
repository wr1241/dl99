package main

import (
	"context"
	"dl99"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	serverHost = flag.String("host", "0.0.0.0", "the server host")
	serverPort = flag.Int("port", 9999, "the server port")
	maxPlayers = flag.Int("max-players", dl99.DefaultMaxPlayers, "max players")
	maxGames   = flag.Int("max-games", dl99.DefaultMaxGames, "max game")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := dl99.NewServer(*maxPlayers, *maxGames)
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				log.Printf("cleaned %d finished games\n", srv.CleanUpFinishedGame())
			case <-ctx.Done():
				log.Println("exit")
				return
			}
		}
	}()

	r := gin.Default()

	// new player
	r.POST("/player", func(c *gin.Context) {
		if playerId, err := srv.NewPlayer(c.PostForm("name")); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"player_id": playerId,
			})
		}
	})

	// new game
	r.POST("/game", func(c *gin.Context) {
		if gameId, err := srv.NewGame(c.PostForm("name")); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"game_id": gameId,
			})
		}
	})

	// list games
	r.GET("/games", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"games": srv.GameBriefs(),
		})
	})

	// join game
	r.POST("/join/:game_id/:player_id", func(c *gin.Context) {
		if err := srv.JoinGame(c.Param("game_id"), c.Param("player_id")); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	// leave game
	r.POST("/leave/:game_id/:player_id", func(c *gin.Context) {
		if err := srv.LeaveGame(c.Param("game_id"), c.Param("player_id")); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	// start game
	r.POST("/start_game", func(c *gin.Context) {
		gameId, ok := c.GetPostForm("game_id")
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("missing game_id"))
			return
		}

		playerId, ok := c.GetPostForm("player_id")
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("missing player_id"))
			return
		}

		if err := srv.StartGame(gameId, playerId); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	// game info
	r.GET("/game/:game_id", func(c *gin.Context) {
		gameDetail, err := srv.GameInfo(c.Param("game_id"))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gameDetail)
	})

	// player info
	r.GET("/player/:player_id", func(c *gin.Context) {
		playerDetail, err := srv.PlayerInfo(c.Param("player_id"))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, playerDetail)
	})

	// play card
	r.POST("/play/:game_id/:player_id/:card_index", func(c *gin.Context) {
		cardIndex, err := strconv.ParseInt(c.Param("card_index"), 10, 32)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid card_index"))
			return
		}
		var cardOption dl99.CardOption
		if err := c.ShouldBind(&cardOption); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := srv.PlayCard(c.Param("game_id"), c.Param("player_id"), int(cardIndex), &cardOption); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	if err := r.Run(fmt.Sprintf("%s:%d", *serverHost, *serverPort)); err != nil {
		log.Println(err)
	}
}
