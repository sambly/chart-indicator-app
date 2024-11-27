package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
)

func getChart(c *gin.Context) {

	q, err := quote.NewQuoteFromYahoo("BTC-USD", "2023-01-01", "2023-12-31", quote.Daily, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(q.CSV())

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"EntryJS":   entryJS,
		"ChartData": q,
	})
}
