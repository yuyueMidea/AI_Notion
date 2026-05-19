package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type LayoutConfig struct {
	SeatCodePrefix  string            `json:"seat_code_prefix"`
	SeatNumberStart int               `json:"seat_number_start"`
	SeatNumberWidth int               `json:"seat_number_width"`
	Zones           []ZoneConfig      `json:"zones"`
	FixedOwnerMap   map[string]string `json:"fixed_owner_map"`
}

type ZoneConfig struct {
	ZoneCode   string `json:"zone_code"`
	ZoneName   string `json:"zone_name"`
	SeatCount  int    `json:"seat_count"`
	FixedCount int    `json:"fixed_count"`
}

func LoadLayoutConfig(path string) (LayoutConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LayoutConfig{}, fmt.Errorf("read layout config %q: %w", path, err)
	}

	var cfg LayoutConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return LayoutConfig{}, fmt.Errorf("parse layout config: %w", err)
	}

	if cfg.SeatCodePrefix == "" {
		cfg.SeatCodePrefix = "A-"
	}
	if cfg.SeatNumberStart <= 0 {
		cfg.SeatNumberStart = 1
	}
	if cfg.SeatNumberWidth <= 0 {
		cfg.SeatNumberWidth = 3
	}
	if cfg.FixedOwnerMap == nil {
		cfg.FixedOwnerMap = map[string]string{}
	}
	if len(cfg.Zones) == 0 {
		return LayoutConfig{}, fmt.Errorf("layout config zones cannot be empty")
	}

	for _, zone := range cfg.Zones {
		if zone.ZoneCode == "" {
			return LayoutConfig{}, fmt.Errorf("zone_code cannot be empty")
		}
		if zone.ZoneName == "" {
			return LayoutConfig{}, fmt.Errorf("zone_name cannot be empty for zone %s", zone.ZoneCode)
		}
		if zone.SeatCount <= 0 {
			return LayoutConfig{}, fmt.Errorf("seat_count must be > 0 for zone %s", zone.ZoneCode)
		}
		if zone.FixedCount < 0 {
			return LayoutConfig{}, fmt.Errorf("fixed_count cannot be < 0 for zone %s", zone.ZoneCode)
		}
		if zone.FixedCount > zone.SeatCount {
			return LayoutConfig{}, fmt.Errorf("fixed_count cannot exceed seat_count for zone %s", zone.ZoneCode)
		}
	}

	return cfg, nil
}
