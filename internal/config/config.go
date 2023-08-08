package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Db     DBConfig     `yaml:"db"`
}

type ServerConfig struct {
	Port             string        `yaml:"port"`
	BotToken         string        `yaml:"botToken" json:"botToken"`
	ReadTimeout      time.Duration `json:"read-timeout" yaml:"read-timeout"`
	WriteTimeout     time.Duration `json:"write-timeout" yaml:"write-timeout"`
	GracefulShutdown time.Duration `json:"graceful-shutdown" yaml:"graceful-shutdown"`
	Cors             struct {
		AllowAll bool     `yaml:"allowAll"`
		Origin   []string `yaml:"origin"`
	} `yaml:"cors"`
	Auth struct {
		JWT struct {
			Realm      string        `json:"realm" yaml:"realm"`
			Key        string        `json:"key" yaml:"key"`
			Timeout    time.Duration `json:"timeout" yaml:"timeout"`
			MaxRefresh time.Duration `json:"max-refresh" yaml:"max-refresh"`
		} `json:"jwt" yaml:"jwt"`
	} `json:"auth" yaml:"auth"`
}

type DBConfig struct {
	DataSourceName string `yaml:"dataSourceName"`
	//Database string `yaml:"database"`
}

func LoadConfig() (Config, error) {
	vp := viper.New()
	var config Config
	vp.AddConfigPath("./././configs")
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")

	if err := vp.ReadInConfig(); err != nil {
		return Config{}, err
	}

	if err := vp.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil

}
