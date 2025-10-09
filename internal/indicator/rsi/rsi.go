package indicatorrsi

import (
	"fmt"
	"main/internal/model"
	"math"

	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type RSI struct {
	*Config
	SignalBuyPoints  []model.IndicatorData
	SignalSellPoints []model.IndicatorData
	RSIValues        []float64
	EMAValues        []float64
}

func NewRSI() (*RSI, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &RSI{
		Config:           cfg,
		SignalBuyPoints:  make([]model.IndicatorData, 0),
		SignalSellPoints: make([]model.IndicatorData, 0),
		RSIValues:        make([]float64, 0),
		EMAValues:        make([]float64, 0),
	}, nil
}

func (s *RSI) Execute(candles quote.Quote) (signalBuyOnLast, signalSellOnLast bool) {
	// --- Очистка сигналов ---
	s.SignalBuyPoints = s.SignalBuyPoints[:0]
	s.SignalSellPoints = s.SignalSellPoints[:0]

	closes := candles.Close
	times := candles.Date
	n := len(closes)

	// --- Минимальное количество баров ---
	minBars := maxInt(s.RSILength, s.EMASlowLength, 20) + 10
	if n < minBars {
		fmt.Println("[RSI] Недостаточно баров для анализа")
		return false, false
	}

	// --- Индикаторы ---
	rsi := talib.Rsi(closes, s.RSILength)
	emaSlow := talib.Ema(closes, s.EMASlowLength)
	emaFast := talib.Ema(closes, 20)

	// Сохраняем для анализа
	s.RSIValues = rsi
	s.EMAValues = emaSlow

	startIndex := maxInt(s.RSILength, s.EMASlowLength, 20)
	lastSignalIndex := -9999
	minBarsBetweenTrades := s.MinBarsBetweenTrades

	for i := startIndex; i < n; i++ {
		if math.IsNaN(rsi[i]) || math.IsNaN(emaSlow[i]) || math.IsNaN(emaFast[i]) {
			continue
		}

		currClose := closes[i]
		prevClose := closes[i-1]
		currEMA := emaSlow[i]
		prevEMA := emaSlow[i-1]
		currRSI := rsi[i]
		prevRSI := rsi[i-1]
		currEMAFast := emaFast[i]

		dateStr := times[i].Format("2006-01-02 15:04")

		// --- BUY CONDITIONS ---
		buyCond1 := prevClose < prevEMA && currClose > currEMA
		buyCond2 := currRSI < s.RSIBuyLevel || (prevRSI < s.RSIBuyLevel && currRSI > prevRSI)
		buyCond3 := currEMAFast > currEMA
		buyCond4 := i-lastSignalIndex >= minBarsBetweenTrades

		buySignal := buyCond1 && buyCond2 && buyCond3 && buyCond4

		if buySignal {
			lastSignalIndex = i
			s.SignalBuyPoints = append(s.SignalBuyPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalBuyOnLast = true
			}

			fmt.Printf("[BUY] %s | Close=%.2f | RSI=%.1f | EMA=%.2f | FastEMA=%.2f\n", dateStr, currClose, currRSI, currEMA, currEMAFast)
			fmt.Printf("      cond1(cross up EMA)=%v cond2(RSI zone)=%v cond3(Fast>Slow)=%v cond4(cooldown)=%v\n",
				buyCond1, buyCond2, buyCond3, buyCond4)
			continue
		}

		// --- SELL CONDITIONS ---
		sellCond1 := prevClose > prevEMA && currClose < currEMA
		sellCond2 := currRSI > s.RSIExitLevel && currRSI < prevRSI
		sellCond3 := i-lastSignalIndex >= minBarsBetweenTrades

		trueCount := 0
		if sellCond1 {
			trueCount++
		}
		if sellCond2 {
			trueCount++
		}
		if sellCond3 {
			trueCount++
		}

		sellSignal := trueCount >= 2

		if sellSignal {
			lastSignalIndex = i
			s.SignalSellPoints = append(s.SignalSellPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalSellOnLast = true
			}

			fmt.Printf("[SELL] %s | Close=%.2f | RSI=%.1f | EMA=%.2f | FastEMA=%.2f\n", dateStr, currClose, currRSI, currEMA, currEMAFast)
			fmt.Printf("       cond1(cross down EMA)=%v cond2(RSI down)=%v cond3(cooldown)=%v  -> matched=%d\n",
				sellCond1, sellCond2, sellCond3, trueCount)
		}
	}

	return signalBuyOnLast, signalSellOnLast
}

// Дополнительные методы для анализа
func (s *RSI) GetLastRSI() float64 {
	if len(s.RSIValues) == 0 {
		return 0
	}
	return s.RSIValues[len(s.RSIValues)-1]
}

func (s *RSI) GetLastEMA() float64 {
	if len(s.EMAValues) == 0 {
		return 0
	}
	return s.EMAValues[len(s.EMAValues)-1]
}

// --- helpers ---
func maxInt(vals ...int) int {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
