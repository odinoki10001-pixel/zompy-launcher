package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	DataDir       string `json:"dataDir"`
	JavaPath      string `json:"javaPath"`
	GameDir       string `json:"gameDir"`
	PacksDir      string `json:"packsDir"`
	SavesDir      string `json:"savesDir"`
	ModsCache     string `json:"modsCache"`
	Ram           int    `json:"ram"` // GB
	Resolution    string `json:"resolution"`
	AutoUpdate    bool   `json:"autoUpdate"`
	CurseForgeKey string `json:"curseforgeKey"`
	AccentColor   string `json:"accentColor"`
	Username      string `json:"username"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".zompy")
	return &Config{
		DataDir:     base,
		GameDir:     filepath.Join(base, "minecraft"),
		PacksDir:    filepath.Join(base, "packs"),
		SavesDir:    filepath.Join(base, "saves"),
		ModsCache:   filepath.Join(base, "mods-cache"),
		JavaPath:    "",
		Ram:         6,
		Resolution:  "1280x720",
		AutoUpdate:  true,
		AccentColor: "#7c8cff",
		Username:    "Steve",
	}
}

func path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zompy", "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(path())
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(path()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o644)
}

func OSName() string   { return runtime.GOOS }
func ArchName() string { return runtime.GOARCH }
