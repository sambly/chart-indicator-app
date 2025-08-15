package sma

import (
	"fmt"
	"main/internal/app"
	"main/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type Handler struct {
	app *app.App
}

func New(app *app.App) *Handler {
	return &Handler{
		app: app,
	}
}

func (h *Handler) Register(router gin.IRouter) {
	router.GET("/sma", h.GetSMA)
}

func (h *Handler) GetSMA(c *gin.Context) {

	a := h.app

	q, err := quote.NewQuoteFromCoinbase("BTC-USD", "2023-01-01", "2023-12-31", quote.Daily)
	if err != nil {
		fmt.Println(err)
		return
	}

	//  Simple Moving Average (SMA)
	sma := talib.Sma(q.Close, 10)
	var smaIndicator []model.IndicatorData
	for index := range q.Close {
		smaIndicator = append(smaIndicator, model.IndicatorData{Date: q.Date[index], Value: sma[index]})
	}

	c.HTML(http.StatusOK, "sma.tmpl", gin.H{
		"EntryJS":      a.EntryJS,
		"ChartData":    q,
		"SMAIndicator": smaIndicator,
	})
}
