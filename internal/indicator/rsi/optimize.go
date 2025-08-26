package indicatorrsi

import (
	"fmt"
	"main/internal/model"
	"math"
	"time"

	"github.com/markcheno/go-quote"
)

type OptimizationResult struct {
	Config          *Config
	Profit          float64
	Trades          int
	WinRate         float64
	Drawdown        float64
	WinRatePercent  float64
	CountSignalBuy  int
	CountSignalSell int
	EquityCurve     []float64
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

func EvaluateRSIStrategy(s *RSI, candles quote.Quote) OptimizationResult {
	profit, trades, winRate, drawdown, equity, countSignalBuy, countSignalSell := backtestRSI(s, candles, true, false)

	return OptimizationResult{
		Config:          s.Config,
		Profit:          profit,
		Trades:          trades,
		WinRate:         winRate,
		Drawdown:        drawdown,
		WinRatePercent:  winRate * 100,
		EquityCurve:     equity,
		CountSignalBuy:  countSignalBuy,
		CountSignalSell: countSignalSell,
	}
}

func OptimizeRSIStrategy(candles quote.Quote) OptimizationResult {
	// Диапазоны параметров
	ranges := map[string][2]int{
		"RSILength":     {7, 21},
		"EMASlowLength": {30, 200},
	}
	floatRanges := map[string][2]float64{
		"RSIBuyLevel":  {20, 40},
		"RSIExitLevel": {60, 80},
	}

	// Шаги
	steps := map[string]int{
		"RSILength":     2,
		"EMASlowLength": 10,
	}
	floatSteps := map[string]float64{
		"RSIBuyLevel":  2,
		"RSIExitLevel": 2,
	}

	best := OptimizationResult{Profit: -math.MaxFloat64}

	// Перебор параметров
	for rsiLen := ranges["RSILength"][0]; rsiLen <= ranges["RSILength"][1]; rsiLen += steps["RSILength"] {
		for emaSlow := ranges["EMASlowLength"][0]; emaSlow <= ranges["EMASlowLength"][1]; emaSlow += steps["EMASlowLength"] {
			for buyLevel := floatRanges["RSIBuyLevel"][0]; buyLevel <= floatRanges["RSIBuyLevel"][1]; buyLevel += floatSteps["RSIBuyLevel"] {
				for exitLevel := floatRanges["RSIExitLevel"][0]; exitLevel <= floatRanges["RSIExitLevel"][1]; exitLevel += floatSteps["RSIExitLevel"] {

					// создаём стратегию
					strat, _ := NewRSI()
					strat.RSILength = rsiLen
					strat.EMASlowLength = emaSlow
					strat.RSIBuyLevel = buyLevel
					strat.RSIExitLevel = exitLevel

					profit, trades, winRate, drawdown, _, _, _ := backtestRSI(strat, candles, false, false)

					// выбираем лучшую стратегию по прибыли
					if profit > best.Profit {
						best = OptimizationResult{
							Config:         strat.Config,
							Profit:         profit,
							Trades:         trades,
							WinRate:        winRate,
							Drawdown:       drawdown,
							WinRatePercent: winRate * 100,
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

func backtestRSI(s *RSI, candles quote.Quote, verbose bool, closeAllAtEnd bool) (
	profit float64,
	trades int,
	winRate float64,
	maxDD float64,
	equityCurve []float64,
	countSignalBuy int,
	CountSignalSell int,
) {
	var positions []*position
	var wins int
	var peak float64
	var currentEquity float64

	// Пересчитаем сигналы по стратегии
	s.Execute(candles)

	equityCurve = make([]float64, len(candles.Close))

	if verbose {
		fmt.Println("=== ДЕТАЛИ СДЕЛОК ===")
		fmt.Printf("%-20s | %-8s | %-10s | %-10s | %-10s | %-8s\n",
			"Время", "Тип", "Цена входа", "Цена выхода", "Прибыль", "Статус")
		fmt.Println("---------------------|----------|------------|------------|------------|----------")
	}

	for i, price := range candles.Close {
		buyNow := containsTime(s.SignalBuyPoints, candles.Date[i])
		sellNow := containsTime(s.SignalSellPoints, candles.Date[i])

		// --- ВХОД (при сигнале BUY - открываем новую позицию) ---
		if buyNow {
			newPos := &position{
				entryPrice: price,
				entryTime:  candles.Date[i],
			}
			positions = append(positions, newPos)

			if verbose {
				fmt.Printf("%-20s | %-8s | %-10.2f | %-10s | %-10s | %-8s\n",
					candles.Date[i].Format("2006-01-02 15:04:05"),
					"BUY",
					price,
					"-",
					"-",
					"OPEN")
			}
		}

		// --- ВЫХОД (при сигнале SELL - закрываем ВСЕ открытые позиции) ---
		if sellNow && len(positions) > 0 {
			for _, pos := range positions {
				pnl := price - pos.entryPrice
				currentEquity += pnl

				status := "LOSS"
				if pnl > 0 {
					wins++
					status = "WIN"
				}

				if verbose {
					fmt.Printf("%-20s | %-8s | %-10.2f | %-10.2f | %-10.2f | %-8s\n",
						candles.Date[i].Format("2006-01-02 15:04:05"),
						"SELL",
						pos.entryPrice,
						price,
						pnl,
						status)
				}

				trades++
			}
			positions = nil
		}

		// --- equity расчет ---
		unrealized := 0.0
		for _, pos := range positions {
			unrealized += price - pos.entryPrice
		}
		equityCurve[i] = currentEquity + unrealized

		// --- обновляем max drawdown ---
		if equityCurve[i] > peak {
			peak = equityCurve[i]
		}
		if dd := peak - equityCurve[i]; dd > maxDD {
			maxDD = dd
		}
	}

	// Закрыть все оставшиеся позиции только если closeAllAtEnd = true
	if closeAllAtEnd && len(positions) > 0 {
		finalPrice := candles.Close[len(candles.Close)-1]
		finalTime := candles.Date[len(candles.Date)-1]

		for _, pos := range positions {
			pnl := finalPrice - pos.entryPrice
			currentEquity += pnl

			status := "LOSS"
			if pnl > 0 {
				wins++
				status = "WIN"
			}

			if verbose {
				fmt.Printf("%-20s | %-8s | %-10.2f | %-10.2f | %-10.2f | %-8s\n",
					finalTime.Format("2006-01-02 15:04:05"),
					"SELL*",
					pos.entryPrice,
					finalPrice,
					pnl,
					status)
			}

			trades++
		}
		equityCurve[len(equityCurve)-1] = currentEquity
	}

	profit = currentEquity

	// win rate
	if trades > 0 {
		winRate = float64(wins) / float64(trades)
	}

	if verbose {
		fmt.Println("\n=== ИТОГОВАЯ СТАТИСТИКА ===")
		fmt.Printf("Всего сделок: %d\n", trades)
		fmt.Printf("Прибыльных: %d (%.1f%%)\n", wins, winRate*100)
		fmt.Printf("Общая прибыль: %.2f\n", profit)
		fmt.Printf("Макс. просадка: %.2f\n", maxDD)
		fmt.Printf("Win Rate: %.2f%%\n", winRate*100)
	}

	return profit, trades, winRate, maxDD, equityCurve, len(s.SignalBuyPoints), len(s.SignalSellPoints)
}

func containsTime(data []model.IndicatorData, t time.Time) bool {
	for _, d := range data {
		if d.Date.Equal(t) {
			return true
		}
	}
	return false
}
