package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type PhysicsConfig struct {
	Gravity           float64 `json:"gravity"`
	Stiffness         float64 `json:"stiffness"`
	StiffnessNear     float64 `json:"stiffness_near"`
	RestDensity       float64 `json:"rest_density"`
	Viscosity         float64 `json:"viscosity"`
	Damping           float64 `json:"damping"`
	InteractionRad    float64 `json:"interaction_rad"`
	InteractionRadSq  float64 `json:"-"`
	InvInteractionRad float64 `json:"-"`
	SpawnCount        int     `json:"spawn_count"`
	PaletteName       string  `json:"palette"`
	PaletteIdx        int     `json:"-"` // Runtime only
	IsPaused          bool    `json:"is_paused"`
}

func (c *PhysicsConfig) UpdateDerived() {
	c.InteractionRadSq = c.InteractionRad * c.InteractionRad
	if c.InteractionRad != 0 {
		c.InvInteractionRad = 1.0 / c.InteractionRad
	} else {
		c.InvInteractionRad = 0
	}
}

type HexPalette struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

type AppConfig struct {
	Presets  map[string]PhysicsConfig `json:"presets"`
	Palettes []HexPalette             `json:"color_palettes"`
}

func LoadSettings(path string) (*AppConfig, error) {
	file, err := os.ReadFile(path)

	if os.IsNotExist(err) {
		fmt.Printf("Config file not found at '%s'. Creating default configuration...\n", path)
		defaultConfig := NewDefaultConfig()
		if saveErr := SaveSettings(path, defaultConfig); saveErr != nil {
			return nil, fmt.Errorf("failed to create default config file: %v", saveErr)
		}
		return defaultConfig, nil
	} else if err != nil {
		return nil, err
	}

	var appConfig AppConfig
	if err := json.Unmarshal(file, &appConfig); err != nil {
		return nil, err
	}

	for k, v := range appConfig.Presets {
		v.UpdateDerived()
		appConfig.Presets[k] = v
	}

	return &appConfig, nil
}

func SaveSettings(path string, appConfig *AppConfig) error {
	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func GetSortedPresetNames(configs map[string]PhysicsConfig) []string {
	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func NewDefaultConfig() *AppConfig {
	cfg := &AppConfig{
		Presets: map[string]PhysicsConfig{
			"Default": {
				Gravity:        0.05,
				Stiffness:      0.07,
				StiffnessNear:  0.08,
				RestDensity:    3.0,
				Viscosity:      0.02,
				Damping:        0.91,
				InteractionRad: 3,
				SpawnCount:     20,
				PaletteName:    "Water",
				IsPaused:       false,
			},
			"Magma": {
				Gravity:        0.03,
				Stiffness:      0.06,
				StiffnessNear:  0.08,
				RestDensity:    6,
				Viscosity:      0.005,
				Damping:        0.96,
				InteractionRad: 3,
				SpawnCount:     20,
				PaletteName:    "Magma",
				IsPaused:       false,
			},
		},
		Palettes: []HexPalette{
			{
				Name: "Water",
				Colors: []string{
					"#A0D8EF", "#4DB4EB", "#008BD2", "#0060B0", "#00388C",
				},
			},
			{
				Name: "Magma",
				Colors: []string{
					"#FFE066", "#FFB300", "#FF4500", "#CC1100", "#660000",
				},
			},
			{
				Name: "Tea",
				Colors: []string{
					"#8B7D6B", "#CDB38B", "#8B4500", "#382201", "#110900",
				},
			},
			{
				Name: "Beer",
				Colors: []string{
					"#FFFFFF", "#FFF44F", "#FFC300", "#D98200", "#6E3B00",
				},
			},
		},
	}

	for k, v := range cfg.Presets {
		v.UpdateDerived()
		cfg.Presets[k] = v
	}

	return cfg
}
