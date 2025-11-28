package gui

import (
	"log"

	g "github.com/AllenDang/giu"
	"github.com/Kondou727/maimai-recommend-intl/internal/config"
)

func Loop(cfg *config.ApiConfig) {
	g.SingleWindow().Layout(
		g.TabBar().TabItems(
			g.TabItem("Load Scores").Layout(LoadScoreView(cfg)),
		),
	)
}

func Run(cfg *config.ApiConfig) {
	log.Print("Running gui...")
	w := g.NewMasterWindow("maimai-recommend-intl", 1280, 720, 0)
	w.Run(func() { Loop(cfg) })
}
