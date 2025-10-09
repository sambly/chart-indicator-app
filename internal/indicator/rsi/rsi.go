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

func (s *RSI) Execute(candles quote.Quote, verbose bool) (signalBuyOnLast, signalSellOnLast bool) {
	// --- Очистка сигналов ---
	s.SignalBuyPoints = s.SignalBuyPoints[:0]
	s.SignalSellPoints = s.SignalSellPoints[:0]

	if len(candles.Close) == 0 {
		fmt.Println("[RSI] candles.Close = 0")
		return false, false
	}

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
	lastBuyIndex := -9999
	lastSellIndex := -9999

	for i := startIndex; i < n; i++ {
		if i >= len(rsi) || i >= len(emaSlow) || i >= len(emaFast) {
			continue
		}
		if math.IsNaN(rsi[i]) || math.IsNaN(rsi[i-1]) ||
			math.IsNaN(emaSlow[i]) || math.IsNaN(emaSlow[i-1]) ||
			math.IsNaN(emaFast[i]) {
			continue
		}
		if math.IsInf(rsi[i], 0) || math.IsInf(emaSlow[i], 0) || math.IsInf(emaFast[i], 0) {
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
		buyCond1 := prevClose < prevEMA && currClose > currEMA         // пересечение ценой медленной EMA снизу вверх (трендовый сигнал)
		buyCond2 := prevRSI < s.RSIBuyLevel && currRSI > s.RSIBuyLevel // RSI пересекает уровень покупки снизу вверх → фильтр импульса (моментум)
		buyCond3 := currEMAFast > currEMA                              // быстрая EMA выше медленной
		buyCond4 := i-lastBuyIndex >= s.MinBarsBetweenTrades           //минимальное расстояние между покупками

		buySignal := buyCond1 && buyCond2 && buyCond3 && buyCond4

		if buySignal {
			lastBuyIndex = i
			s.SignalBuyPoints = append(s.SignalBuyPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalBuyOnLast = true
			}

			if verbose {
				fmt.Printf("[BUY] %s | Close=%.2f | RSI=%.1f | EMA=%.2f | FastEMA=%.2f\n", dateStr, currClose, currRSI, currEMA, currEMAFast)
				fmt.Printf("      cond1(cross up EMA)=%v cond2(RSI zone)=%v cond3(Fast>Slow)=%v cond4(cooldown)=%v\n",
					buyCond1, buyCond2, buyCond3, buyCond4)
			}
			continue
		}

		// --- SELL CONDITIONS ---
		sellCond1 := prevClose > prevEMA && currClose < currEMA           // цена пересекла EMA сверху вниз → сигнал разворота тренда
		sellCond2 := prevRSI > s.RSIExitLevel && currRSI < s.RSIExitLevel //RSI пересёк уровень выхода сверху вниз → сигнал ослабления импульса
		sellCond3 := i-lastSellIndex >= s.MinBarsBetweenTrades            //защита от слишком частых сигналов
		sellCond4 := currEMAFast < currEMA                                // быстрая EMA ниже медленной

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
		if sellCond4 {
			trueCount++
		}

		sellSignal := trueCount >= s.CountSellSignals

		if sellSignal {
			lastSellIndex = i
			s.SignalSellPoints = append(s.SignalSellPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalSellOnLast = true
			}

			if verbose {
				fmt.Printf("[SELL] %s | Close=%.2f | RSI=%.1f | EMA=%.2f | FastEMA=%.2f\n", dateStr, currClose, currRSI, currEMA, currEMAFast)
				fmt.Printf("       cond1(cross down EMA)=%v cond2(RSI down)=%v cond3(cooldown)=%v\n",
					sellCond1, sellCond2, sellCond3)
			}
		}
	}

	return signalBuyOnLast, signalSellOnLast
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
