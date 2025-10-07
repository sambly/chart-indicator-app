package app

import (
	"time"

	"github.com/markcheno/go-quote"
)

var (
	SymbolDefault    = "BTC-USD"
	IntervalDefault  = quote.Min60
	EndDateDefault   = time.Now().Format("2006-01-02")                   // Текущая дата
	StartDateDefault = time.Now().AddDate(0, -1, 0).Format("2006-01-02") // Дата месяц назад от текущей
)

type Indicator struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type App struct {
	Symbol         string                                  `json:"symbol"`
	StartDate      string                                  `json:"start_date"`
	EndDate        string                                  `json:"end_date"`
	IntervalString string                                  `json:"interval"`
	Interval       quote.Period                            `json:"-"`
	Quote          map[string]map[quote.Period]quote.Quote `json:"-"`
}

func NewApp() *App {
	return &App{
		Quote:     make(map[string]map[quote.Period]quote.Quote),
		Symbol:    SymbolDefault,
		StartDate: StartDateDefault,
		EndDate:   EndDateDefault,
		Interval:  IntervalDefault,
	}
}
