package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/deb-sig/bill-file-converter/core"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MinerU MinerUConfig `yaml:"mineru"`
}

type MinerUConfig struct {
	BaseURL     string            `yaml:"base_url"`
	LangList    []string          `yaml:"lang_list"`
	Backend     string            `yaml:"backend"`
	ParseMethod string            `yaml:"parse_method"`
	Timeout     string            `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
	Headers     map[string]string `yaml:"headers"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func DefaultConfig() Config {
	return Config{
		MinerU: MinerUConfig{
			BaseURL:     "http://127.0.0.1:<port>",
			LangList:    []string{"ch"},
			Backend:     "hybrid-auto-engine",
			ParseMethod: "auto",
			Timeout:     "10m",
			MaxRetries:  1,
			Headers:     map[string]string{},
		},
	}
}

func (c Config) MinerUHTTPConfig() (core.MinerUHTTPConfig, error) {
	timeout := 10 * time.Minute
	if c.MinerU.Timeout != "" {
		parsed, err := time.ParseDuration(c.MinerU.Timeout)
		if err != nil {
			return core.MinerUHTTPConfig{}, fmt.Errorf("mineru.timeout: %w", err)
		}
		timeout = parsed
	}
	langList := c.MinerU.LangList
	if len(langList) == 0 {
		langList = []string{"ch"}
	}
	backend := c.MinerU.Backend
	if backend == "" {
		backend = "hybrid-auto-engine"
	}
	parseMethod := c.MinerU.ParseMethod
	if parseMethod == "" {
		parseMethod = "auto"
	}
	return core.MinerUHTTPConfig{
		BaseURL:     c.MinerU.BaseURL,
		LangList:    langList,
		Backend:     backend,
		ParseMethod: parseMethod,
		Timeout:     timeout,
		MaxRetries:  c.MinerU.MaxRetries,
		Headers:     c.MinerU.Headers,
	}, nil
}

func WriteDefaultConfig(path string) error {
	config := DefaultConfig()
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
