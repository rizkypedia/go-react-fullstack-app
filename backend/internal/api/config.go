package api

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	JWTSecret      []byte
	DemoUsername   string
	DemoPassword   string
	FrontendOrigin string
}

func LoadConfig() (Config, error) {
	if os.Getenv("APP_ENV") != "development" {
		return Config{}, errors.New("this starter only supports APP_ENV=development; configure a production identity provider before deployment")
	}

	config := Config{
		JWTSecret:      []byte(os.Getenv("JWT_SECRET")),
		DemoUsername:   os.Getenv("DEMO_USERNAME"),
		DemoPassword:   os.Getenv("DEMO_PASSWORD"),
		FrontendOrigin: os.Getenv("FRONTEND_ORIGIN"),
	}
	if len(config.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if config.DemoUsername == "" || config.DemoPassword == "" {
		return Config{}, errors.New("DEMO_USERNAME and DEMO_PASSWORD must be set")
	}
	if config.FrontendOrigin == "" {
		return Config{}, errors.New("FRONTEND_ORIGIN must be set to the exact frontend origin")
	}
	origin, err := url.Parse(config.FrontendOrigin)
	if err != nil || origin.Scheme == "" || origin.Host == "" ||
		(origin.Scheme != "http" && origin.Scheme != "https") ||
		origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return Config{}, fmt.Errorf("FRONTEND_ORIGIN must be an exact HTTP(S) origin")
	}
	if strings.TrimSpace(config.FrontendOrigin) != config.FrontendOrigin {
		return Config{}, errors.New("FRONTEND_ORIGIN must not contain leading or trailing whitespace")
	}
	return config, nil
}
