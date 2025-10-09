package feeder

import (
	"github.com/markcheno/go-quote"
)

type Feeder interface {
	GetQuote(symbol, startDate, endDate string, period quote.Period) (quote.Quote, error)
}

type FeederApiCoinbase struct{}

type FeederJSONFile struct{}

func NewFeederApiCoinbase() *FeederApiCoinbase {
	return &FeederApiCoinbase{}
}

func NewFeederJSONFile() *FeederJSONFile {
	return &FeederJSONFile{}
}

func (f *FeederApiCoinbase) GetQuote(symbol, startDate, endDate string, period quote.Period) (quote.Quote, error) {
	q, err := quote.NewQuoteFromCoinbase(symbol, startDate, endDate, period)
	return q, err
}

func (f *FeederJSONFile) GetQuote(symbol, startDate, endDate string, period quote.Period) (quote.Quote, error) {
	q, err := quote.NewQuoteFromJSONFile("./example/BTC-USD.json")
	return q, err
}
