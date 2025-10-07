package indicatorrsi

import (
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
	router.GET("rsi/default-data", h.GetTrendRSIDefault)
	router.POST("rsi/update", h.UpdateTrendRSIData)
	router.POST("rsi/apply-config", h.ApplyRSIConfig)
	router.POST("rsi/save-config", h.SaveRSIConfig)
	router.GET("rsi/default-config", h.GetRSIDefaultConfig)
	router.POST("rsi/optimize", h.OptimizeRSIStrategy)
	router.POST("rsi/evaluate", h.EvaluateRSIStrategyHandler)
}

func (h *Handler) GetTrendRSIDefault(c *gin.Context) {

	a := h.app
	intervalQuote := utils.ParsePeriod(string(a.Interval))
	c.JSON(http.StatusOK, gin.H{
		"symbol":           a.Symbol,
		"startDate":        a.StartDate,
		"endDate":          a.EndDate,
		"interval":         intervalQuote,
		"config":           h.rsi.Config,
		"currentOpti":      h.currentOpti,
		"optimization":     h.optimization,
		"chartData":        a.Quote[a.Symbol][intervalQuote],
		"signalBuyPoints":  h.rsi.SignalBuyPoints,
		"signalSellPoints": h.rsi.SignalSellPoints,
	})
}

func (h *Handler) UpdateTrendRSIData(c *gin.Context) {
	a := h.app

	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.Interval = utils.ParsePeriod(a.IntervalString)

	// Получаем новые котировки
	q, err := quote.NewQuoteFromCoinbase(a.Symbol, a.StartDate, a.EndDate, a.Interval)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if a.Quote[a.Symbol] == nil {
		a.Quote[a.Symbol] = make(map[quote.Period]quote.Quote)
	}
	a.Quote[a.Symbol][a.Interval] = q

	// Выполняем RSI
	h.rsi.Execute(q)

	c.JSON(http.StatusOK, gin.H{
		"symbol":           a.Symbol,
		"startDate":        a.StartDate,
		"endDate":          a.EndDate,
		"interval":         a.Interval,
		"config":           h.rsi.Config,
		"currentOpti":      h.currentOpti,
		"optimization":     h.optimization,
		"chartData":        a.Quote[a.Symbol][a.Interval],
		"signalBuyPoints":  h.rsi.SignalBuyPoints,
		"signalSellPoints": h.rsi.SignalSellPoints,
	})
}

func (h *Handler) ApplyRSIConfig(c *gin.Context) {
	a := h.app

	if err := c.ShouldBindJSON(&h.rsi.Config); err != nil {
		println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	q, ok := a.Quote[a.Symbol][a.Interval]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no quote data for symbol/interval"})
		return
	}

	h.rsi.Execute(q)

	c.JSON(http.StatusOK, gin.H{
		"chartData":        a.Quote[a.Symbol][a.Interval],
		"signalBuyPoints":  h.rsi.SignalBuyPoints,
		"signalSellPoints": h.rsi.SignalSellPoints,
	})
}

func (h *Handler) SaveRSIConfig(c *gin.Context) {
	if err := c.ShouldBindJSON(&h.rsi.Config); err != nil {
		println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.rsi.Config.SaveConfig(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) GetRSIDefaultConfig(c *gin.Context) {
	cfg, err := NewConfig()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.rsi.Config = cfg

	c.JSON(http.StatusOK, gin.H{
		"config": h.rsi.Config,
	})
}

func (h *Handler) OptimizeRSIStrategy(c *gin.Context) {
	a := h.app

	q, ok := a.Quote[a.Symbol][a.Interval]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no quote data for symbol/interval"})
		return
	}

	optimizationResult := OptimizeRSIStrategy(q)
	h.rsi.Config = optimizationResult.Config
	h.optimization = optimizationResult
	h.rsi.Execute(q)

	c.JSON(http.StatusOK, gin.H{
		"config":           h.rsi.Config,
		"optimization":     optimizationResult,
		"chartData":        a.Quote[a.Symbol][a.Interval],
		"signalBuyPoints":  h.rsi.SignalBuyPoints,
		"signalSellPoints": h.rsi.SignalSellPoints,
	})
}

func (h *Handler) EvaluateRSIStrategyHandler(c *gin.Context) {
	a := h.app

	q, ok := a.Quote[a.Symbol][a.Interval]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no quote data for symbol/interval"})
		return
	}

	optimizationResult := EvaluateRSIStrategy(h.rsi, q)
	h.currentOpti = optimizationResult

	c.JSON(http.StatusOK, gin.H{
		"currentOpti": h.currentOpti,
	})
}
