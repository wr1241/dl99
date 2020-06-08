package main

import (
	"dl99"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var (
	serverHost = flag.String("host", "0.0.0.0", "the server host")
	serverPort = flag.Int("port", 9999, "the server port")

	rng     = rand.New(rand.NewSource(time.Now().UnixNano()))
	players = make(map[uuid.UUID]*dl99.Player)
	games   = make(map[uuid.UUID]*dl99.Game)
)

func main() {
	flag.Parse()

	//POST /player?name=xxxxxx
	//GET /player?id=0
	http.HandleFunc("/player", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			name := req.PostFormValue("name")
			if len(name) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				log.Println("no name")
				return
			}

			var id uuid.UUID
			for {
				id = uuid.New()
				if _, ok := players[id]; !ok {
					players[id] = dl99.NewPlayer(id, name)
					data, err := json.Marshal(players[id])
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						log.Printf("marshal player failed: %v\n", err)
						return
					}
					if _, err := w.Write(data); err != nil {
						log.Printf("marshal player failed: %v\n", err)
						return
					}
					break
				}
			}
		case http.MethodGet:
			id, err := uuid.Parse(req.URL.Query().Get("id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid player id: %s\n", req.URL.Query().Get("id"))
				return
			}
			if player, ok := players[id]; ok {
				data, err := json.Marshal(player)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("marshal player failed: %v\n", err)
					return
				}
				if _, err := w.Write(data); err != nil {
					log.Printf("send player failed: %v\n", err)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("player %d not found\n", id)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("player interface allow POST or GET")
			return
		}
	})

	//POST /game
	//GET /game?id=0
	http.HandleFunc("/game", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			name := req.PostFormValue("name")
			if len(name) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				log.Println("no name")
				return
			}

			var id uuid.UUID
			for {
				id = uuid.New()
				if _, ok := games[id]; !ok {
					games[id] = dl99.NewGame(id, name)
					data, err := json.Marshal(games[id])
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						log.Printf("marshal game failed: %v\n", err)
						return
					}
					if _, err := w.Write(data); err != nil {
						log.Printf("marshal game failed: %v\n", err)
						return
					}
					break
				}
			}
		case http.MethodGet:
			id, err := uuid.Parse(req.URL.Query().Get("id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid game id: %s\n", req.URL.Query().Get("id"))
				return
			}

			if game, ok := games[id]; ok {
				data, err := json.Marshal(game)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("marshal game failed: %v\n", err)
					return
				}
				if _, err := w.Write(data); err != nil {
					log.Printf("send game failed: %v\n", err)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("game %d not found\n", id)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("game interface allow POST or GET")
			return
		}
	})

	// GET /games
	http.HandleFunc("/games", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			data, err := json.Marshal(games)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("marshal games faield: %v\n", err)
				return
			}
			if _, err := w.Write(data); err != nil {
				log.Printf("send games faile: %v\n", err)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("games interface allow GET")
			return
		}
	})

	// POST /join_game?game_id=0&player_id=0
	http.HandleFunc("/join_game", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			gameId, err := uuid.Parse(req.PostFormValue("game_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid game id: %s\n", req.PostFormValue("game_id"))
				return
			}
			game, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("game %d not found", gameId)
				return
			}

			playerId, err := uuid.Parse(req.PostFormValue("player_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid player id: %s\n", req.PostFormValue("player_id"))
				return
			}
			player, ok := players[playerId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("player %d not found", playerId)
				return
			}

			if err := game.PlayerJoin(player); err != nil {
				w.WriteHeader(http.StatusForbidden)
				log.Printf("player join failed: %v\n", err)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("join_game interface allow POST")
			return
		}
	})

	// POST /start_game?game_id=0
	http.HandleFunc("/start_game", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			id, err := uuid.Parse(req.PostFormValue("game_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid game id: %s\n", req.PostFormValue("game_id"))
				return
			}
			if game, ok := games[id]; ok {
				if err := game.StartGame(); err != nil {
					w.WriteHeader(http.StatusForbidden)
					log.Printf("start game failed: %v\n", err)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("game %d not found\n", id)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("start_game interface allow POST")
			return
		}
	})

	// POST /play?game_id=0&player_id=0&card_index=0&card_options={}
	http.HandleFunc("/play", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			gameId, err := uuid.Parse(req.PostFormValue("game_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid game id: %s\n", req.PostFormValue("game_id"))
				return
			}
			game, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("game %d not found\n", gameId)
				return
			}

			playerId, err := uuid.Parse(req.PostFormValue("player_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid player id: %s\n", req.PostFormValue("player_id"))
				return
			}
			player, ok := players[playerId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("player %d not found", playerId)
				return
			}

			cardIndex, err := strconv.ParseInt(req.PostFormValue("card_index"), 10, 32)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("invalid card_index %s\n", req.PostFormValue("card_index"))
				return
			}

			var cardOption *dl99.CardOption
			if len(req.PostFormValue("card_option")) > 0 {
				if err := json.Unmarshal([]byte(req.PostFormValue("card_option")), cardOption); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					log.Printf("invalid card_option: %s\n", req.PostFormValue("card_option"))
					return
				}
			}

			if err := player.Play(game, int(cardIndex), cardOption); err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				log.Printf("play failed: %v\n", err)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Println("play interface allow POST")
			return
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *serverHost, *serverPort), nil); err != nil {
		log.Fatal(err)
	}
}
