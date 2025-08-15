package model

import "time"

type IndicatorData struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}
