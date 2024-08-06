package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	HTTP       HTTPConfig
	Collectors []CollectorConfig
}

type HTTPConfig struct {
	ListenAddr string
}

type CollectorConfig struct {
	Name   string
	URL    string
	Auth   AuthConfig
	Fields []FieldConfig
}

type AuthMethod string

const (
	AuthMethodBearer AuthMethod = "bearer"
	AuthMethodXToken AuthMethod = "x-token"
)

type AuthConfig struct {
	Method AuthMethod
	Token  string `json:"-"`
}

type FieldConfig struct {
	Name        string
	Description string
	Query       string
}

func Load(file string) (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()
	fmt.Println(v.GetEnvPrefix())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if file != "" {
		v.SetConfigFile(file)
	}

	v.SetConfigType("yml")

	if err := v.MergeInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	for i := range c.Collectors {
		key := fmt.Sprintf("COLLECTORS_%s_AUTH_TOKEN", strings.ToUpper(c.Collectors[i].Name))
		if v.Get(key) != nil && v.Get(key).(string) != "" {
			c.Collectors[i].Auth.Token = v.Get(key).(string)
		}
	}

	return &c, nil
}
