package main

import (
	"encoding/json"
	"errors"
	"github.com/freeeve/uci"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type EngineWrapper struct {
	Engine *uci.Engine
	LastAccessed time.Time
}

func main() {
	os.Mkdir("./data", 0755)
	os.Mkdir("./data/syzygy", 0755)

	http.HandleFunc("/move", ChessServer)
	http.ListenAndServe(":8081", nil)
}

func ChessServer(w http.ResponseWriter, r *http.Request) {
	var responseObj interface{}

	w.Header().Set("Content-Type", "application/json")

	if game, ok := r.URL.Query()["game"];ok {
		if fenString, ok := r.URL.Query()["fen"]; ok {
			elo := 800

			var eloString []string

			if eloString, ok = r.URL.Query()["elo"]; ok {
				eloTemp, err := strconv.Atoi(eloString[0])
				if err == nil {
					elo = eloTemp
				}
			}

			result, err := GetStockfishResults(strings.Join(game, " "), fenString[0], elo)

			if err == nil {
				responseObj = result
			} else {
				// Try to kill the engine and restart
				result, err = GetStockfishResults(strings.Join(game, " "), fenString[0], elo)

				if err == nil {
					responseObj = result
				} else {
					responseObj = err
				}
			}
		} else {
			responseObj = errors.New("fen parameter is missing")
		}
	} else {
		responseObj = errors.New("game parameter is missing")
	}

	if errObj, ok := responseObj.(error); ok {
		responseObj = map[string]string {
			"errorMessage": errObj.Error(),
		}
	}

	bytes, err := json.Marshal(responseObj)

	if err != nil {
		log.Fatal(err)
	} else {
		w.WriteHeader(200)
		w.Write(bytes)
	}
}

var engines map[string]EngineWrapper = map[string]EngineWrapper{}

func getSurvivalTime() time.Duration {
	survivalSeconds := os.Getenv("SURVIVAL_TIME")
	result, err := strconv.Atoi(survivalSeconds)
	if err != nil {
		result = 30
	}

	return time.Duration(result) * time.Second
}

func getMaxEngines() int {
	maxEngines := os.Getenv("MAX_ENGINES")
	result, err := strconv.Atoi(maxEngines)
	if err != nil {
		result = 200
	}

	return result
}

func pruneEngines() {
	for len(engines) > getMaxEngines() {
		var gameId *string
		var engineToPrune *EngineWrapper

		for k, v := range engines {
			if engineToPrune == nil {
				engineToPrune = &v
				gameId = &k
			} else {
				if v.LastAccessed.Before(engineToPrune.LastAccessed) {
					engineToPrune = &v
					gameId = &k
				}
			}
		}

		delete(engines, *gameId)
	}
}

func GetEngine(gameID string) (engine *uci.Engine, err error) {
	if wrapper, ok := engines[gameID]; ok {
		wrapper.LastAccessed = time.Now()
		engine = wrapper.Engine
	} else {
		pruneEngines()

		engine, err = uci.NewEngine(os.Getenv("STOCKFISH_PATH"))
		if err == nil {
			// set some engine options
			engine.SetOptions(uci.Options{
				Hash:    1024,
				Ponder:  false,
				OwnBook: true,
				MultiPV: 2,
				Threads: 2,
			})

			wrapper := EngineWrapper{
				Engine:       engine,
				LastAccessed: time.Now(),
			}
			engines[gameID] = wrapper
		}
	}

	go func(gameID string) {
		time.Sleep(getSurvivalTime() * 2)
		if wrapper, ok := engines[gameID]; ok {
			if wrapper.LastAccessed.Add(getSurvivalTime()).Before(time.Now()) {
				delete(engines, gameID)
				wrapper.Engine.Close()
			}
		}
	}(gameID)

	return engine, err
}

func GetStockfishResults(gameID string, fenString string, elo int) (result *uci.Results, err error) {
	// Based on this article.  Though, we could create a test with different configurations to run
	// Stockfish against Stockfish to determine new Elos
	// http://www.talkchess.com/forum3/viewtopic.php?t=69731
	/*
	skillLevelElos := []int{
		1231,
		1341,
		1443,
		1538,
		1678,
		1823,
		1881,
		1976,
		2067,
		2129,
		2221,
		2320,
		2406,
		2483,
		2571,
		2657,
		2761,
		2815,
		2872,
		2905,
		3450,
	}

	 */

	eng, err := GetEngine(gameID)
	if err == nil {
		// set the starting position
		eng.SetFEN(fenString)

		/**
		skillLevel := 0
		for i, skillLevelElo := range skillLevelElos {
			if skillLevelElo < elo {
				skillLevel = i
			}
		}
		*/

		//eng.SendOption("Skill Level", skillLevel)
		eng.SendOption("UCI_LimitStrength", true)
		eng.SendOption("UCI_Elo", elo)
		eng.SendOption("SyzygyPath", "./data/syzygy")

		// set some result filter options
		result, err = eng.Go(5, "", 10000, uci.HighestDepthOnly, uci.IncludeUpperbounds, uci.IncludeLowerbounds)
	}

	return result, err
}