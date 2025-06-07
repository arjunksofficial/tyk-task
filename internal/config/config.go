package config

import "github.com/spf13/viper"

type Config struct {
	App struct {
		Name string `json:"name"`
		Port string `json:"port"`
	} `json:"app"`
	Routes []struct {
		Path string `json:"path"`
		Host string `json:"host"`
	} `json:"routes"`
	Redis struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		DB       int    `json:"db"`
		Password string `json:"password"`
	} `json:"redis"`
}

var cfg *Config

func (c *Config) GetPort() string {
	if c.App.Port == "" {
		return "9000" // Default port if not set
	}
	return c.App.Port
}

func (c *Config) GetRoutes() []struct {
	Path string `json:"path"`
	Host string `json:"host"`
} {
	if c.Routes == nil {
		return []struct {
			Path string `json:"path"`
			Host string `json:"host"`
		}{}
	}
	return c.Routes
}
func (c *Config) GetRedisConfig() struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	DB       int    `json:"db"`
	Password string `json:"password"`
} {
	return c.Redis
}

// config is under cmd/apigw/config/<env>/master.yaml

func ReadConfig() (*Config, error) {
	viper.SetConfigName("master")         // name of config file (without extension)
	viper.AddConfigPath("./config/local") // path to look for the config file in
	viper.SetConfigType("yaml")           // or viper.SetConfigType("json") for JSON files
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetConfig() *Config {
	if cfg == nil {
		var err error
		cfg, err = ReadConfig()
		if err != nil {
			panic("Failed to read config: " + err.Error())
		}
	}
	return cfg
}
