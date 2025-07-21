package indicator

import "time"

type Indicator struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}
