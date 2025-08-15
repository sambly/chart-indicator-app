package trendSniper

import (
	"encoding/json"
	"fmt"
	"main/internal/app"
	"main/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
)

type Handler struct {
	app    *app.App
	sniper *Sniper
}

func New(app *app.App) (*Handler, error) {
	sniper, err := NewSniper()
	if err != nil {
		return nil, err
	}
	return &Handler{
		app:    app,
		sniper: sniper,
	}, nil
}

func (h *Handler) Register(router gin.IRouter) {
	router.GET("/sniper", h.GetTrendSniper)
}

func (h *Handler) GetTrendSniper(c *gin.Context) {

	a := h.app
	symbol := c.DefaultQuery("symbol", a.Symbol)
	startDate := c.DefaultQuery("start_date", a.StartDate)
	endDate := c.DefaultQuery("end_date", a.EndDate)
	interval := c.DefaultQuery("interval", string(a.Interval))
	intervalQuote := utils.ParsePeriod(interval)

	if err := c.Bind(h.sniper.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configJSON, err := json.Marshal(h.sniper.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	formType := c.Query("form_type")
	switch formType {

	case "update":
		q, err := quote.NewQuoteFromCoinbase(symbol, startDate, endDate, intervalQuote)
		if err != nil {
			fmt.Println(err)
			return
		}

		if a.Quote[symbol] == nil {
			a.Quote[symbol] = make(map[quote.Period]quote.Quote)
		}
		a.Quote[symbol][intervalQuote] = q
		h.sniper.RMITrendSniper(q)

	case "config":
		h.sniper.RMITrendSniper(a.Quote[symbol][intervalQuote])
	case "save":
		if err := h.sniper.Config.SaveConfig(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case "optimize":
		optimizationResult := OptimizeSniper(a.Quote[symbol][intervalQuote])
		h.sniper.Config = optimizationResult.Config

		configJSON, err = json.Marshal(h.sniper.Config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.sniper.RMITrendSniper(a.Quote[symbol][intervalQuote])

	default:

		if a.Quote[symbol] == nil {
			a.Quote[symbol] = make(map[quote.Period]quote.Quote)
		}

		// Первая загрузка страницы или просто обновление
		if _, exists := a.Quote[symbol][intervalQuote]; !exists {
			q, err := quote.NewQuoteFromCoinbase(symbol, startDate, endDate, intervalQuote)
			if err != nil {
				fmt.Println(err)
				return
			}
			// q := quote.Quote{}

			a.Quote[symbol][intervalQuote] = q
			h.sniper.RMITrendSniper(q)
		}
	}

	c.HTML(http.StatusOK, "sniper.tmpl", gin.H{
		"EntryJS":          a.EntryJS,
		"EntryCSS":         a.EntryCSS,
		"ChartData":        a.Quote[symbol][intervalQuote],
		"SignalBuyPoints":  h.sniper.SignalBuyPoints,
		"SignalSellPoints": h.sniper.SignalSellPoints,
		"Config":           string(configJSON),
		"FormValues": gin.H{
			"symbol":     symbol,
			"start_date": startDate,
			"end_date":   endDate,
			"interval":   interval,
		},
	})
}
