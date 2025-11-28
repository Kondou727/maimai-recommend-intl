package gui

import (
	"log"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	"github.com/Kondou727/maimai-recommend-intl/internal/app"
	"github.com/Kondou727/maimai-recommend-intl/internal/config"
	"github.com/sqweek/dialog"
)

var tsv string

func LoadScoreView(cfg *config.ApiConfig) giu.Layout {
	return g.Layout{
		g.Button("Import TSV file").OnClick(func() {
			importTSV(cfg)
		}),
	}
}

func importTSV(cfg *config.ApiConfig) {
	filename, err := dialog.File().Filter("TSV files", "tsv").Load()
	if err == dialog.ErrCancelled {
		log.Printf("File dialog cancelled by user")
		return
	}
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
	err = app.ImportTSVToDB(filename, cfg)
	if err != nil {
		log.Printf("Failed to import TSV to db: %s", err)
		return
	}
	log.Printf("Success!")
	dialog.Message("TSV import success ^_^").Info()
}
