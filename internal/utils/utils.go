package utils

import (
	"strings"

	"github.com/markcheno/go-quote"
)

func ParsePeriod(input string) quote.Period {
	input = strings.ToLower(input)

	switch input {
	case "60":
		return quote.Min1
	case "3m":
		return quote.Min3
	case "300":
		return quote.Min5
	case "900":
		return quote.Min15
	case "1800":
		return quote.Min30
	case "3600":
		return quote.Min60
	case "2h":
		return quote.Hour2
	case "4h":
		return quote.Hour4
	case "6h":
		return quote.Hour6
	case "8h":
		return quote.Hour8
	case "12h":
		return quote.Hour12
	case "d":
		return quote.Daily
	case "3d":
		return quote.Day3
	case "w":
		return quote.Weekly
	case "m":
		return quote.Monthly
	default:
		return quote.Min60
	}
}
