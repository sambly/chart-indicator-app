package extremum

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
	router.GET("/extremum", h.GetHighLow)
}

func (h *Handler) GetHighLow(c *gin.Context) {

	a := h.app

	q, err := quote.NewQuoteFromCoinbase("BTC-USD", "2023-01-01", "2023-12-31", quote.Daily)
	if err != nil {
		fmt.Println(err)
		return
	}

	high := talib.Max(q.High, 10)
	low := talib.Min(q.Low, 10)

	var highIndicator []model.IndicatorData
	var lowIndicator []model.IndicatorData

	for index := range q.Close {
		highIndicator = append(highIndicator, model.IndicatorData{Date: q.Date[index], Value: high[index]})
		lowIndicator = append(lowIndicator, model.IndicatorData{Date: q.Date[index], Value: low[index]})
	}

	c.HTML(http.StatusOK, "extremum.tmpl", gin.H{
		"EntryJS":       a.EntryJS,
		"ChartData":     q,
		"HighIndicator": highIndicator,
		"LowIndicator":  lowIndicator,
	})
}
