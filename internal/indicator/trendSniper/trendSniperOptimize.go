package trendSniper

import (
	"fmt"
	"math"
	"time"

	"github.com/markcheno/go-quote"
)

type OptimizationResult struct {
	Config   *Config
	Profit   float64
	Trades   int
	WinRate  float64
	Drawdown float64
}

func (o OptimizationResult) String() string {
	return fmt.Sprintf(
		"=== Optimization Result ===\n"+
			"Config: %+v\n"+
			"Profit: %.2f\n"+
			"Trades: %d\n"+
			"WinRate: %.2f%%\n"+
			"Drawdown: %.2f\n",
		o.Config, o.Profit, o.Trades, o.WinRate*100, o.Drawdown,
	)
}

// OptimizeSniper runs a grid search over parameter ranges to find the best configuration
func OptimizeSniper(candles quote.Quote) OptimizationResult {

	// Диапазоны параметров для оптимизации
	ranges := map[string][2]float64{
		"RSIMFIBuyLevel":  {30, 45},
		"RSIMFIExitLevel": {45, 60},
		"EMAFastLength":   {5, 15},
		"ATRMultiplier":   {0.3, 1.0},
	}
	steps := map[string]float64{
		"RSIMFIBuyLevel":  1,
		"RSIMFIExitLevel": 1,
		"EMAFastLength":   1,
		"ATRMultiplier":   0.1,
	}

	best := OptimizationResult{Profit: -math.MaxFloat64}

	for buyLevel := ranges["RSIMFIBuyLevel"][0]; buyLevel <= ranges["RSIMFIBuyLevel"][1]; buyLevel += steps["RSIMFIBuyLevel"] {
		for exitLevel := ranges["RSIMFIExitLevel"][0]; exitLevel <= ranges["RSIMFIExitLevel"][1]; exitLevel += steps["RSIMFIExitLevel"] {
			for emaFast := int(ranges["EMAFastLength"][0]); emaFast <= int(ranges["EMAFastLength"][1]); emaFast += int(steps["EMAFastLength"]) {
				for atrMult := ranges["ATRMultiplier"][0]; atrMult <= ranges["ATRMultiplier"][1]; atrMult += steps["ATRMultiplier"] {

					// TODO сделать Err
					sniper, _ := NewSniper()
					sniper.RSIMFIBuyLevel = buyLevel
					sniper.RSIMFIExitLevel = exitLevel
					sniper.EMAFastLength = emaFast
					sniper.ATRMultiplier = atrMult

					profit, trades, winRate, drawdown := backtest(sniper, candles)

					if profit > best.Profit {
						best = OptimizationResult{
							Config:   sniper.Config,
							Profit:   profit,
							Trades:   trades,
							WinRate:  winRate,
							Drawdown: drawdown,
						}
					}
				}
			}
		}
	}
	return best
}

// backtest runs a simple backtest on given candles
type position struct {
	entryPrice float64
	entryTime  time.Time
}

func backtest(s *Sniper, candles quote.Quote) (profit float64, trades int, winRate float64, maxDD float64) {
	var pos *position
	var wins int
	peak := 0.0
	s.RMITrendSniper(candles)

	for i := range candles.Close {
		price := candles.Close[i]
		buyNow := containsTime(s.SignalBuyPoints, candles.Date[i])
		sellNow := containsTime(s.SignalSellPoints, candles.Date[i])

		if buyNow && pos == nil {
			pos = &position{entryPrice: price, entryTime: candles.Date[i]}
			trades++
		}
		if sellNow && pos != nil {
			pnl := price - pos.entryPrice
			profit += pnl
			if pnl > 0 {
				wins++
			}
			pos = nil
		}

		if profit > peak {
			peak = profit
		}
		if dd := peak - profit; dd > maxDD {
			maxDD = dd
		}
	}
	if trades > 0 {
		winRate = float64(wins) / float64(trades)
	}
	return
}
