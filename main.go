package main

import (
	"log"

	app "github.com/Kondou727/maimai-recommend-intl/internal/app"
	config "github.com/Kondou727/maimai-recommend-intl/internal/config"
	scoresdb "github.com/Kondou727/maimai-recommend-intl/internal/database/scores"
	songdatadb "github.com/Kondou727/maimai-recommend-intl/internal/database/songdata"
	_ "modernc.org/sqlite"
)

func main() {
	log.Printf("Starting maimai-recommend-intl")

	scoresDB, err := app.LoadScoresDB()
	if err != nil {
		log.Fatalf("Failed loading scores DB: %s", err)
	}
	defer scoresDB.Close()

	songdataDB, err := app.LoadSongdataDB()
	if err != nil {
		log.Fatalf("Failed loading song data DB: %s", err)
	}
	defer songdataDB.Close()

	scoresDBQueries := scoresdb.New(scoresDB)
	songdataDBQueries := songdatadb.New(songdataDB)

	cfg := config.ApiConfig{
		ScoresDB:          scoresDB,
		ScoresDBQueries:   scoresDBQueries,
		SongdataDB:        songdataDB,
		SongdataDBQueries: songdataDBQueries,
	}
	/*
		err = cfg.loadTSV()
		if err != nil {
			log.Fatal(err)
		}

	*/
	err = app.PopulateSongData(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = app.PullJackets(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = app.PopulateSongDataCN(&cfg)
	if err != nil {
		log.Fatal(err)
	}

}
