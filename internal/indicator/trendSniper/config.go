package trendSniper

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	primaryPath  = "internal/indicator/trendSniper/config.yaml"
	fallbackPath = "internal/indicator/trendSniper/config.default.yaml"
)

type Config struct {
	// Core lengths
	RSILength       int `yaml:"rsi_length"`
	EMAFastLength   int `yaml:"ema_fast_length"`
	EMASlowLength   int `yaml:"ema_slow_length"`
	ATRLength       int `yaml:"atr_length"`
	ATRSMALength    int `yaml:"atr_sma_length"`
	VolumeSMALength int `yaml:"volume_sma_length"`

	// Levels and filters
	RSIMFIBuyLevel   float64 `yaml:"rsi_mfi_buy_level"`  // Cross up to enter
	RSIMFIExitLevel  float64 `yaml:"rsi_mfi_exit_level"` // Cross down to exit
	EMAMinDelta      float64 `yaml:"ema_min_delta"`      // Minimal EMA fast slope (per bar)
	ATRMultiplier    float64 `yaml:"atr_multiplier"`     // Multiplier for ATR average filter
	BuyVolumeFactor  float64 `yaml:"buy_volume_factor"`  // Multiplier for average volume on buys
	SellVolumeFactor float64 `yaml:"sell_volume_factor"` // Multiplier for average volume on sells

	// Behavior
	DeduplicateSignals bool `yaml:"deduplicate_signals"` // If true, avoids repeated signals while already in the same state
}

func NewConfig() (*Config, error) {
	var config Config

	fileData, err := os.ReadFile(primaryPath)
	if err != nil {
		fileData, err = os.ReadFile(fallbackPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read both config.yaml and config.default.yaml: %w", err)
		}
	}

	if err := yaml.Unmarshal(fileData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func (c *Config) SaveConfig() error {

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(primaryPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
