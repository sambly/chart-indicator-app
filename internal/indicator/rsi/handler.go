package indicatorrsi

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
	app          *app.App
	rsi          *RSI
	currentOpti  OptimizationResult
	optimization OptimizationResult
}

func New(app *app.App) (*Handler, error) {
	rsi, err := NewRSI()
	if err != nil {
		return nil, err
	}
	return &Handler{
		app: app,
		rsi: rsi,
	}, nil
}

func (h *Handler) Register(router gin.IRouter) {
	router.GET("/rsi", h.GetTrendRSI)
}

func (h *Handler) GetTrendRSI(c *gin.Context) {

	a := h.app
	symbol := c.DefaultQuery("symbol", a.Symbol)
	startDate := c.DefaultQuery("start_date", a.StartDate)
	endDate := c.DefaultQuery("end_date", a.EndDate)
	interval := c.DefaultQuery("interval", string(a.Interval))
	intervalQuote := utils.ParsePeriod(interval)

	if err := c.Bind(h.rsi.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	formType := c.Query("form_type")
	switch formType {

	case "update":

		a.Symbol = symbol
		a.StartDate = startDate
		a.EndDate = endDate
		a.Interval = intervalQuote

		q, err := quote.NewQuoteFromCoinbase(symbol, startDate, endDate, intervalQuote)
		if err != nil {
			fmt.Println(err)
			return
		}

		if a.Quote[symbol] == nil {
			a.Quote[symbol] = make(map[quote.Period]quote.Quote)
		}
		a.Quote[symbol][intervalQuote] = q
		h.rsi.Execute(q)

	case "config":
		h.rsi.Execute(a.Quote[symbol][intervalQuote])
	case "save":
		if err := h.rsi.Config.SaveConfig(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case "get_default_config":

		cfg, err := NewConfig()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.rsi.Config = cfg

	case "optimize":
		optimizationResult := OptimizeRSIStrategy(a.Quote[symbol][intervalQuote])
		h.rsi.Config = optimizationResult.Config
		h.optimization = optimizationResult
		h.rsi.Execute(a.Quote[symbol][intervalQuote])

	case "currentOptimize":
		optimizationResult := EvaluateRSIStrategy(h.rsi, a.Quote[symbol][intervalQuote])
		h.currentOpti = optimizationResult

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

			a.Quote[symbol][intervalQuote] = q
			h.rsi.Execute(q)
		}
	}

	configJSON, err := json.Marshal(h.rsi.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "rsi.tmpl", gin.H{
		"EntryJS":             a.EntryJS,
		"EntryCSS":            a.EntryCSS,
		"ChartData":           a.Quote[symbol][intervalQuote],
		"SignalBuyPoints":     h.rsi.SignalBuyPoints,
		"SignalSellPoints":    h.rsi.SignalSellPoints,
		"Config":              string(configJSON),
		"CurrentOptimization": h.currentOpti,
		"Optimization":        h.optimization,
		"FormValues": gin.H{
			"symbol":     symbol,
			"start_date": startDate,
			"end_date":   endDate,
			"interval":   interval,
		},
	})
}
