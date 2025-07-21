package trendSniper

import (
	"log"
	indicator "main/Indicator"

	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

// Константы параметров индикатора
const (
	RSILength       = 14
	EMAFastLength   = 8
	EMASlowLength   = 50
	ATRLength       = 14
	ATRSMALength    = 14
	VolumeSMALength = 20

	RSIMFIBuyLevel   = 62.0
	RSIMFIExitLevel  = 38.0
	EMAMinDelta      = 0.0001
	ATRMultiplier    = 0.5
	BuyVolumeFactor  = 1.2
	SellVolumeFactor = 0.8
)

type Indicator struct {
	SignalBuyPoints  []indicator.Indicator
	SignalSellPoints []indicator.Indicator
}

func NewIndicator() *Indicator {
	return &Indicator{
		SignalBuyPoints:  make([]indicator.Indicator, 0),
		SignalSellPoints: make([]indicator.Indicator, 0),
	}
}

func (in *Indicator) RMITrendSniper(candles quote.Quote) (signalBuy, signalSell bool) {
	closes := candles.Close
	highs := candles.High
	lows := candles.Low
	volumes := candles.Volume
	times := candles.Date

	// Проверка минимального количества данных
	requiredLength := maxCandles(
		RSILength+1,
		EMAFastLength+1,
		EMASlowLength,
		ATRLength+ATRSMALength,
		VolumeSMALength,
	) + 1

	if len(closes) < requiredLength {
		log.Printf("Not enough candles: have %d, need at least %d", len(closes), requiredLength)
		return false, false
	}

	// Расчёт RSI и MFI
	rsi := talib.Rsi(closes, RSILength)
	mfi := talib.Mfi(highs, lows, closes, volumes, RSILength)

	// Усреднение RSI и MFI
	minLen := min(len(rsi), len(mfi))
	rsiMfi := make([]float64, minLen)
	for j := 0; j < minLen; j++ {
		rsiMfi[j] = (rsi[j] + mfi[j]) / 2
	}

	// EMA
	emaFast := talib.Ema(closes, EMAFastLength)
	emaSlow := talib.Ema(closes, EMASlowLength)

	// Изменение EMA Fast
	emaFastChange := make([]float64, len(emaFast))
	for j := 1; j < len(emaFast); j++ {
		emaFastChange[j] = emaFast[j] - emaFast[j-1]
	}

	// ATR и его сглаживание
	atr := talib.Atr(highs, lows, closes, ATRLength)
	atrAvg := talib.Sma(atr, ATRSMALength)
	atrFilter := atrAvg[len(atrAvg)-1] * ATRMultiplier

	// Средний объём
	volumeAvg := talib.Sma(volumes, VolumeSMALength)

	var lastBuy, lastSell bool

	for idx := requiredLength; idx < len(closes); idx++ {
		if idx >= len(rsiMfi) || idx-1 >= len(rsiMfi) ||
			idx >= len(emaFastChange) || idx >= len(emaSlow) ||
			idx >= len(atr) || idx >= len(volumeAvg) || idx >= len(times) {
			continue
		}

		currentRsiMfi := rsiMfi[idx]
		prevRsiMfi := rsiMfi[idx-1]
		currentEmaFastChange := emaFastChange[idx]
		currentClose := closes[idx]
		currentVolume := volumes[idx]
		currentTime := times[idx]

		signalBuy := prevRsiMfi < RSIMFIBuyLevel &&
			currentRsiMfi > RSIMFIBuyLevel &&
			currentRsiMfi > RSIMFIExitLevel &&
			currentEmaFastChange > EMAMinDelta &&
			currentClose > emaSlow[idx] &&
			atr[idx] > atrFilter &&
			currentVolume > volumeAvg[idx]*BuyVolumeFactor

		signalSell := currentRsiMfi < RSIMFIExitLevel &&
			currentEmaFastChange < -EMAMinDelta &&
			currentClose < emaSlow[idx] &&
			atr[idx] > atrFilter &&
			currentVolume > volumeAvg[idx]*SellVolumeFactor

		if signalBuy {
			in.SignalBuyPoints = append(in.SignalBuyPoints, indicator.Indicator{
				Date:  currentTime,
				Value: currentClose,
			})
			lastBuy = true
		} else {
			lastBuy = false
		}

		if signalSell {
			in.SignalSellPoints = append(in.SignalSellPoints, indicator.Indicator{
				Date:  currentTime,
				Value: currentClose,
			})
			lastSell = true
		} else {
			lastSell = false
		}
	}

	return lastBuy, lastSell
}

// Вспомогательные функции
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxCandles(vals ...int) int {
	if len(vals) == 0 {
		panic("max requires at least one argument")
	}
	maxVal := vals[0]
	for _, v := range vals[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
