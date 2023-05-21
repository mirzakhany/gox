package os

import (
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
)

// LoadFromEnv load and validate env variables into given target.
// example:
//
//		type config struct {
//			 Env      string `env:"ENV,required" envDefault:"local"`
//			 LogLevel string `env:"LOG_LEVEL,required" envDefault:"debug"`
//
//			 HTTPPort string `env:"HTTP_PORT" envDefault:"9091"`
//		}
//
//	    cfg := config{}
//		if err := gox.LoadFromEnv(&cfg); err != nil {
//			 ...
//		}
func LoadFromEnv(config interface{}) error {
	if err := env.Parse(config); err != nil {
		return err
	}
	if err := validator.New().Struct(config); err != nil {
		return err
	}
	return nil
}

// MustGetEnv is using os.LookupEnv to get an env variable.
// it will return def instead if value is not present in env
func MustGetEnv(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
