package trendSniper

import (
	"log"
	"main/internal/model"
	"time"

	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type Sniper struct {
	*Config
	SignalBuyPoints  []model.IndicatorData
	SignalSellPoints []model.IndicatorData
}

func NewSniper() (*Sniper, error) {

	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &Sniper{
		Config:           cfg,
		SignalBuyPoints:  make([]model.IndicatorData, 0),
		SignalSellPoints: make([]model.IndicatorData, 0),
	}, nil
}

// RMITrendSniper calculates signals on the whole series and returns
// whether the *last* bar generated a Buy or Sell signal.
//
// It also fills s.SignalBuyPoints and s.SignalSellPoints for plotting/backtests.
func (s *Sniper) RMITrendSniper(candles quote.Quote) (signalBuyOnLast, signalSellOnLast bool) {
	// Reset points every call
	s.SignalBuyPoints = s.SignalBuyPoints[:0]
	s.SignalSellPoints = s.SignalSellPoints[:0]

	closes := candles.Close
	highs := candles.High
	lows := candles.Low
	volumes := candles.Volume
	times := candles.Date

	if len(closes) == 0 || len(highs) != len(closes) || len(lows) != len(closes) || len(volumes) != len(closes) || len(times) != len(closes) {
		log.Printf("trendSniper: input series lengths mismatch or empty: close=%d high=%d low=%d vol=%d time=%d", len(closes), len(highs), len(lows), len(volumes), len(times))
		return false, false
	}

	// --- Indicators ---
	rsi := talib.Rsi(closes, s.RSILength)
	mfi := talib.Mfi(highs, lows, closes, volumes, s.RSILength)
	emaFast := talib.Ema(closes, s.EMAFastLength)
	emaSlow := talib.Ema(closes, s.EMASlowLength)
	atr := talib.Atr(highs, lows, closes, s.ATRLength)
	atrAvg := talib.Sma(atr, s.ATRSMALength)
	volAvg := talib.Sma(volumes, s.VolumeSMALength)

	// EMA fast slope (simple 1-bar delta)
	emaFastChange := make([]float64, len(emaFast))
	for i := 1; i < len(emaFast); i++ {
		emaFastChange[i] = emaFast[i] - emaFast[i-1]
	}

	// --- Determine the first bar where all indicator values are reliable ---
	// Conservative lookback: ensure all moving averages and ATR-SMA are formed
	startIndex := maxInt(
		s.RSILength,                // RSI/MFI need this many bars
		s.EMAFastLength+1,          // +1 for slope using previous bar
		s.EMASlowLength,            // slow EMA formed
		s.ATRLength+s.ATRSMALength, // ATR then SMA of ATR
		s.VolumeSMALength,          // volume SMA formed
	)
	if startIndex >= len(closes) {
		log.Printf("trendSniper: not enough candles: have %d, need at least %d", len(closes), startIndex+1)
		return false, false
	}

	// --- Iterate and build signals ---
	inPosition := false
	lastIndex := len(closes) - 1

	for i := startIndex; i < len(closes); i++ {
		// Combined oscillator (RSI+MFI)/2
		currR := (rsi[i] + mfi[i]) / 2.0
		prevR := (rsi[i-1] + mfi[i-1]) / 2.0

		// Filters per-bar (dynamic)
		atrFilter := atrAvg[i] * s.ATRMultiplier
		price := closes[i]
		v := volumes[i]

		buy := (prevR < s.RSIMFIBuyLevel && currR >= s.RSIMFIBuyLevel) && // crossing up buy level
			(emaFastChange[i] > s.EMAMinDelta) && // positive momentum
			(price > emaSlow[i]) && // above trend baseline
			(atr[i] > atrFilter) && // sufficient volatility
			(v > volAvg[i]*s.BuyVolumeFactor)

		sell := (prevR > s.RSIMFIExitLevel && currR <= s.RSIMFIExitLevel) && // crossing down exit level
			(emaFastChange[i] < -s.EMAMinDelta) &&
			(price < emaSlow[i]) &&
			(atr[i] > atrFilter) &&
			(v > volAvg[i]*s.SellVolumeFactor)

		// Deduplicate if needed (avoid repeated buys while already "in")
		if s.DeduplicateSignals {
			if inPosition {
				buy = false // don't buy again while in
			} else {
				sell = false // don't sell if not in
			}
		}

		if buy {
			inPosition = true
			s.SignalBuyPoints = append(s.SignalBuyPoints, model.IndicatorData{Date: times[i], Value: price})
			if i == lastIndex {
				signalBuyOnLast = true
			}
		}
		if sell {
			inPosition = false
			s.SignalSellPoints = append(s.SignalSellPoints, model.IndicatorData{Date: times[i], Value: price})
			if i == lastIndex {
				signalSellOnLast = true
			}
		}
	}

	return signalBuyOnLast, signalSellOnLast
}

func containsTime(data []model.IndicatorData, t time.Time) bool {
	for _, d := range data {
		if d.Date.Equal(t) {
			return true
		}
	}
	return false
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
