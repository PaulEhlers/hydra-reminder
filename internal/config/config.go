package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DurationMinutes int    `json:"duration_minutes"`
	AlertColor      string `json:"alert_color"`
	AlertStyle      string `json:"alert_style"` // "color" or "blink"
	HotkeyEnabled   bool   `json:"hotkey_enabled"`
	HotkeyModifiers uint32 `json:"hotkey_modifiers"` // See win32 MOD_ALT, MOD_CONTROL etc
	HotkeyResetKey  uint32 `json:"hotkey_reset_key"` // Virtual key code for reset
	Autostart       bool   `json:"autostart"`
}

func DefaultConfig() *Config {
	return &Config{
		DurationMinutes: 30,
		AlertColor:      "#FF0000",
		AlertStyle:      "color",
		HotkeyEnabled:   false,
		// CTRL + ALT + R
		HotkeyModifiers: 0x0002 | 0x0001, // MOD_CONTROL | MOD_ALT
		HotkeyResetKey:  0x52,            // 'R'
		Autostart:       false,
	}
}

func GetConfigPath() (string, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appData, "HydraReminder")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = Save(cfg)
			return cfg, nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
