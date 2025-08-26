package indicatorrsi

import (
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
	// Reset signals but keep historical data storage
	s.SignalBuyPoints = make([]model.IndicatorData, 0)
	s.SignalSellPoints = make([]model.IndicatorData, 0)
	s.RSIValues = make([]float64, 0)
	s.EMAValues = make([]float64, 0)

	closes := candles.Close
	times := candles.Date

	n := len(closes)
	if n < 50 { // Минимум баров для надежных сигналов
		return false, false
	}

	// --- Indicators ---
	rsi := talib.Rsi(closes, s.RSILength)
	emaSlow := talib.Ema(closes, s.EMASlowLength)
	emaFast := talib.Ema(closes, 20) // Быстрая EMA для определения тренда

	// Определяем стартовый индекс (берем максимальный период)
	startIndex := maxInt(s.RSILength, s.EMASlowLength, 20)
	if startIndex >= n {
		return false, false
	}

	// Сохраняем значения индикаторов для отладки/анализа
	s.RSIValues = rsi
	s.EMAValues = emaSlow

	var lastSignalIndex int = -9999
	minBarsBetweenTrades := s.MinBarsBetweenTrades
	// if minBarsBetweenTrades == 0 {
	// 	minBarsBetweenTrades = 3
	// }

	for i := startIndex; i < n; i++ {
		// Пропускаем NaN значения
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

		// Улучшенные условия входа
		buyConditions := []bool{
			// Основное условие: цена пересекает EMA снизу вверх
			prevClose < prevEMA && currClose > currEMA,
			// RSI в зоне перепроданности И начинает расти
			currRSI < s.RSIBuyLevel && currRSI > prevRSI,
			// Быстрая EMA выше медленной (восходящий тренд)
			currEMAFast > currEMA,
			// Cooldown между сделками
			i-lastSignalIndex >= minBarsBetweenTrades,
			// Объем выше среднего (опционально, если есть данные объема)
		}

		// Улучшенные условия выхода
		sellConditions := []bool{
			// Цена пересекает EMA сверху вниз
			prevClose > prevEMA && currClose < currEMA,
			// RSI в зоне перекупленности И начинает падать
			currRSI > s.RSIExitLevel && currRSI < prevRSI,
			// Cooldown между сделками
			i-lastSignalIndex >= minBarsBetweenTrades,
		}

		// Для входа требуем выполнения всех условий
		buySignal := true
		for _, condition := range buyConditions {
			if !condition {
				buySignal = false
				break
			}
		}

		// Для выхода достаточно большинства условий
		sellSignal := true
		for _, condition := range sellConditions {
			if !condition {
				sellSignal = false
				break
			}
		}

		if buySignal {
			lastSignalIndex = i
			s.SignalBuyPoints = append(s.SignalBuyPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalBuyOnLast = true
			}
		}

		if sellSignal {
			lastSignalIndex = i
			s.SignalSellPoints = append(s.SignalSellPoints, model.IndicatorData{
				Date:  times[i],
				Value: currClose,
			})
			if i == n-1 {
				signalSellOnLast = true
			}
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
