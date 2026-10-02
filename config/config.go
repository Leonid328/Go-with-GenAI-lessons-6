package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	ServerPort    int    `json:"server_port"`
	Environment   string `json:"environment"`
	DatabaseURL   string `json:"database_url,omitempty"`
	DebugMode     bool   `json:"debug_mode"`
	AdminPassword string `json:"-"`
}

type FileStorage interface {
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadCloser, error)
}

type OSStorage struct{}

func (OSStorage) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (OSStorage) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

var storage FileStorage = OSStorage{}

func SaveConfig(path string, cfg Config) error {
	return saveConfig(storage, path, cfg)
}

func LoadConfig(path string) (Config, error) {
	return loadConfig(storage, path)
}

func saveConfig(fs FileStorage, path string, cfg Config) error {
	f, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("save config: create %s: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("save config: encode %s: %w", path, err)
	}
	return nil
}

func loadConfig(fs FileStorage, path string) (Config, error) {
	var cfg Config
	f, err := fs.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("load config: open %s: %w", path, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("load config: decode %s: %w", path, err)
	}
	return cfg, nil
}

func ValidateConfig(cfg Config) error {
	if cfg.ServerPort < 1024 || cfg.ServerPort > 65535 {
		return fmt.Errorf("validate config: port %d out of range 1024-65535", cfg.ServerPort)
	}
	switch cfg.Environment {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("validate config: bad environment %q", cfg.Environment)
	}
	return nil
}
