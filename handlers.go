package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type Indicator struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

func home(c *gin.Context) {
	c.HTML(http.StatusOK, "base.tmpl", gin.H{})
}

func getSMA(c *gin.Context) {

	q, err := quote.NewQuoteFromYahoo("BTC-USD", "2023-01-01", "2023-12-31", quote.Daily, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	//  Simple Moving Average (SMA)
	sma := talib.Sma(q.Close, 10)
	var smaIndicator []Indicator
	for index := range q.Close {
		smaIndicator = append(smaIndicator, Indicator{Date: q.Date[index], Value: sma[index]})
	}

	c.HTML(http.StatusOK, "sma.tmpl", gin.H{
		"EntryJS":      entryJS,
		"ChartData":    q,
		"SMAIndicator": smaIndicator,
	})
}

func getHighLow(c *gin.Context) {

	q, err := quote.NewQuoteFromYahoo("BTC-USD", "2023-01-01", "2023-12-31", quote.Daily, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	high := talib.Max(q.High, 10)
	low := talib.Min(q.Low, 10)

	var highIndicator []Indicator
	var lowIndicator []Indicator

	for index := range q.Close {
		highIndicator = append(highIndicator, Indicator{Date: q.Date[index], Value: high[index]})
		lowIndicator = append(lowIndicator, Indicator{Date: q.Date[index], Value: low[index]})
	}

	c.HTML(http.StatusOK, "extremum.tmpl", gin.H{
		"EntryJS":       entryJS,
		"ChartData":     q,
		"HighIndicator": highIndicator,
		"LowIndicator":  lowIndicator,
	})
}

func plug(c *gin.Context) {
	c.HTML(http.StatusOK, "plug.tmpl", gin.H{
		"EntryJS": entryJS,
		"Data":    "Heloo",
	})
}
