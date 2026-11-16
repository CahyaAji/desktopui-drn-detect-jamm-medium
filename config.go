package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configAppName = "Anti Drone System"

type DetectorConfig struct {
	Address                 string `json:"address"`
	Port                    int    `json:"port"`
	DetectionTimeoutSeconds int    `json:"detectionTimeoutSeconds"`
}

type MapsConfig struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Zoom        float64 `json:"zoom"`
	MapTilerKey string  `json:"maptilerKey"`
}

type JammerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type AppConfig struct {
	Version  int            `json:"version"`
	Detector DetectorConfig `json:"detector"`
	Maps     MapsConfig     `json:"maps"`
	Jammer   JammerConfig   `json:"jammer"`
}

func defaultConfig() AppConfig {
	return AppConfig{
		Version: 1,
		Detector: DetectorConfig{
			Address:                 "192.168.1.100",
			Port:                    7,
			DetectionTimeoutSeconds: 10,
		},
		Maps: MapsConfig{
			Latitude:    -6.2,
			Longitude:   106.816666,
			Zoom:        14,
			MapTilerKey: "",
		},
		Jammer: JammerConfig{
			Address: "192.168.1.150",
			Port:    8000,
		},
	}
}

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(configDir, configAppName, "config.json"), nil
}

func loadOrCreateConfig() (AppConfig, string, error) {
	path, err := configPath()
	if err != nil {
		return AppConfig{}, "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return AppConfig{}, path, fmt.Errorf("create config directory: %w", err)
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		config := defaultConfig()
		if err := saveConfigFile(path, config); err != nil {
			return AppConfig{}, path, err
		}
		return config, path, nil
	}
	if err != nil {
		return AppConfig{}, path, fmt.Errorf("read config file: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return AppConfig{}, path, fmt.Errorf("parse config file: %w", err)
	}

	return config, path, nil
}

func saveConfigFile(path string, config AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*.json")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("set config file permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write config file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close config file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}

	return nil
}
