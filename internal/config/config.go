package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Binance BinanceConfig `mapstructure:"binance"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

type BinanceConfig struct {
	Spot    APIEndpoints `mapstructure:"spot"`
	Futures APIEndpoints `mapstructure:"futures"`
}

type APIEndpoints struct {
	RestURL      string `mapstructure:"restUrl"`
	WebSocketURL string `mapstructure:"websocketUrl"`
}

type LoggingConfig struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	LogRequests  bool   `mapstructure:"logRequests"`
	LogResponses bool   `mapstructure:"logResponses"`
	OutputPath   string `mapstructure:"outputPath"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.readTimeout", "30s")
	v.SetDefault("server.writeTimeout", "30s")
	v.SetDefault("server.shutdownTimeout", "10s")

	v.SetDefault("binance.spot.restUrl", "https://api.binance.com")
	v.SetDefault("binance.spot.websocketUrl", "wss://stream.binance.com:9443")
	v.SetDefault("binance.futures.restUrl", "https://fapi.binance.com")
	v.SetDefault("binance.futures.websocketUrl", "wss://fstream.binance.com")

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.logRequests", true)
	v.SetDefault("logging.logResponses", true)
	v.SetDefault("logging.outputPath", "stdout")

	// Read config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Environment variable overrides
	v.SetEnvPrefix("PROXY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
