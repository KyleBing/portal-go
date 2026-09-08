package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type DatabaseConfig struct {
	Host               string `json:"host"`
	User               string `json:"user"`
	Password           string `json:"password"`
	Port               int    `json:"port"`
	MultipleStatements bool   `json:"multipleStatements"`
	Timezone           string `json:"timezone"`
}

var (
	mu     sync.RWMutex
	cfg    DatabaseConfig
	cfgPath string
)

func IsDocker() bool {
	return os.Getenv("NODE_ENV") == "docker" || os.Getenv("DOCKER_ENV") == "true"
}

func ConfigPath() string {
	if cfgPath != "" {
		return cfgPath
	}
	candidates := []string{
		"config/configDatabase.json",
		filepath.Join("..", "config", "configDatabase.json"),
		filepath.Join("..", "..", "config", "configDatabase.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			cfgPath = c
			return cfgPath
		}
	}
	cfgPath = "config/configDatabase.json"
	return cfgPath
}

func Load() (DatabaseConfig, error) {
	mu.Lock()
	defer mu.Unlock()
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DatabaseConfig{}, err
	}
	var c DatabaseConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return DatabaseConfig{}, err
	}
	if c.Port == 0 {
		c.Port = 3306
	}
	cfg = c
	return cfg, nil
}

func Get() DatabaseConfig {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}

func Save(c DatabaseConfig) error {
	mu.Lock()
	defer mu.Unlock()
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	cfg = c
	return nil
}

func Host() string {
	c := Get()
	if IsDocker() {
		return "mysql"
	}
	return c.Host
}
