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
	EntryJS   string
	EntryCSS  string
	Symbol    string
	StartDate string
	EndDate   string
	Interval  quote.Period
	Quote     map[string]map[quote.Period]quote.Quote
}

func NewApp(entryJS, entryCSS string) *App {
	return &App{
		EntryJS:   entryJS,
		EntryCSS:  entryCSS,
		Quote:     make(map[string]map[quote.Period]quote.Quote),
		Symbol:    SymbolDefault,
		StartDate: StartDateDefault,
		EndDate:   EndDateDefault,
		Interval:  IntervalDefault,
	}
}
