package indicatorrsi

import (
	"fmt"
	"main/internal/model"
	"math"
	"time"

	"github.com/markcheno/go-quote"
)

type OptimizationResult struct {
	Config          *Config   `json:"-"`
	Profit          float64   `json:"profit"`
	Trades          int       `json:"trades"`
	WinRate         float64   `json:"winRate"`
	Drawdown        float64   `json:"drawdown"`
	WinRatePercent  float64   `json:"winRatePercent"`
	CountSignalBuy  int       `json:"countSignalBuy"`
	CountSignalSell int       `json:"countSignalSell"`
	EquityCurve     []float64 `json:"-"`
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
	rsiMin, rsiMax := 7, 21
	emaMin, emaMax := 30, 200

	buyMin, buyMax, buyStep := 20.0, 40.0, 2.0
	exitMin, exitMax, exitStep := 60.0, 80.0, 2.0

	best := OptimizationResult{Profit: -math.MaxFloat64}

	// Перебор параметров (используем целочисленные шаги для float)
	for rsiLen := rsiMin; rsiLen <= rsiMax; rsiLen += 2 {
		for emaSlow := emaMin; emaSlow <= emaMax; emaSlow += 10 {
			for ib := 0; ; ib++ {
				buyLevel := buyMin + float64(ib)*buyStep
				if buyLevel > buyMax+1e-9 {
					break
				}
				for ie := 0; ; ie++ {
					exitLevel := exitMin + float64(ie)*exitStep
					if exitLevel > exitMax+1e-9 {
						break
					}

					strat, err := NewRSI()
					if err != nil {
						// если не удалось создать стратегию, пропускаем
						fmt.Println("NewRSI error:", err)
						continue
					}
					strat.RSILength = rsiLen
					strat.EMASlowLength = emaSlow
					strat.RSIBuyLevel = buyLevel
					strat.RSIExitLevel = exitLevel

					profit, trades, winRate, drawdown, equity, _, _ := backtestRSI(strat, candles, false, false)

					if profit > best.Profit {
						// скопируем Config чтобы не хранить ссылку на временный объект
						cfgCopy := *strat.Config
						best = OptimizationResult{
							Config:         &cfgCopy,
							Profit:         profit,
							Trades:         trades,
							WinRate:        winRate,
							Drawdown:       drawdown,
							WinRatePercent: winRate * 100,
							EquityCurve:    equity,
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
	countSignalSell int,
) {

	var positions []*position
	var wins int
	var peak float64 = 0.0
	var currentEquity float64 = 0.0

	// Пересчитаем сигналы по стратегии
	s.Execute(candles, false)

	// build fast lookup maps
	buyMap := buildSignalMap(s.SignalBuyPoints)
	sellMap := buildSignalMap(s.SignalSellPoints)

	equityCurve = make([]float64, len(candles.Close))

	if verbose {
		fmt.Println("=== ДЕТАЛИ СДЕЛОК ===")
		fmt.Printf("%-20s | %-8s | %-10s | %-10s | %-10s | %-8s\n",
			"Время", "Тип", "Цена входа", "Цена выхода", "Прибыль", "Статус")
		fmt.Println("---------------------|----------|------------|------------|------------|----------")
	}

	for i, price := range candles.Close {
		tKey := candles.Date[i].UnixNano()
		buyNow := buyMap[tKey]
		sellNow := sellMap[tKey]

		if buyNow {
			positions = append(positions, &position{
				entryPrice: price,
				entryTime:  candles.Date[i],
			})
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

		// unrealized
		unrealized := 0.0
		for _, pos := range positions {
			unrealized += price - pos.entryPrice
		}
		equityCurve[i] = currentEquity + unrealized

		// update peak & drawdown (peak starts from 0 initial equity)
		if equityCurve[i] > peak {
			peak = equityCurve[i]
		}
		if dd := peak - equityCurve[i]; dd > maxDD {
			maxDD = dd
		}
	}

	// close remaining positions optionally
	if closeAllAtEnd && len(positions) > 0 {
		finalPrice := candles.Close[len(candles.Close)-1]
		finalTime := candles.Date[len(candles.Date)-1]

		for _, pos := range positions {
			pnl := finalPrice - pos.entryPrice
			currentEquity += pnl
			if pnl > 0 {
				wins++
			}
			trades++
			if verbose {
				fmt.Printf("%-20s | %-8s | %-10.2f | %-10.2f | %-10.2f | %-8s\n",
					finalTime.Format("2006-01-02 15:04:05"),
					"SELL*",
					pos.entryPrice,
					finalPrice,
					pnl,
					func() string {
						if pnl > 0 {
							return "WIN"
						} else {
							return "LOSS"
						}
					}())
			}
		}
		equityCurve[len(equityCurve)-1] = currentEquity
	}

	profit = currentEquity

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

func buildSignalMap(data []model.IndicatorData) map[int64]bool {
	m := make(map[int64]bool, len(data))
	for _, d := range data {
		m[d.Date.UnixNano()] = true
	}
	return m
}
