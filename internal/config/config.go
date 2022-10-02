package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	runAddr         string
	accSystem       string
	dbDSN           string
	sessionLifeTime time.Duration
}
type (
	configOption func(o *configOptions)

	configOptions struct {
		osArgs       []string
		envVars      map[string]string
		ignoreOsArgs bool
	}
)

func withOsArgs(osArgs []string) configOption {
	return func(o *configOptions) {
		o.osArgs = osArgs
	}
}

func WithEnvVars(envVars map[string]string) configOption {
	return func(o *configOptions) {
		o.envVars = envVars
	}
}

func IgnoreOsArgs() configOption {
	return func(o *configOptions) {
		o.ignoreOsArgs = true
	}
}

func New(opts ...configOption) *Config {
	configOptions := &configOptions{
		osArgs: os.Args[1:],
		envVars: map[string]string{
			"RUN_ADDRESS":            os.Getenv("RUN_ADDRESS"),
			"ACCRUAL_SYSTEM_ADDRESS": os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
			"DATABASE_URI":           os.Getenv("DATABASE_URI"),
		},
	}

	for _, o := range opts {
		o(configOptions)
	}

	// default:
	// s := config{"http://localhost:8080", "localhost:8080", os.Getenv("HOME") + "/storage.csv", "user=postgres password=postgres host=localhost port=5432 dbname=testdb"}
	s := Config{runAddr: "localhost:8080", accSystem: "http://localhost:8085", sessionLifeTime: 30 * time.Minute}

	if v := configOptions.envVars["RUN_ADDRESS"]; v != "" {
		s.runAddr = v
	}
	if v := configOptions.envVars["ACCRUAL_SYSTEM_ADDRESS"]; v != "" {
		s.accSystem = v
	}
	if v := configOptions.envVars["DATABASE_URI"]; v != "" {
		s.dbDSN = v
	}

	if !configOptions.ignoreOsArgs {
		fs := flag.NewFlagSet("myFS", flag.ContinueOnError)
		if !fs.Parsed() {
			fs.StringVar(&s.runAddr, "a", s.runAddr, "service address")
			fs.StringVar(&s.accSystem, "r", s.accSystem, "accrual system address")
			fs.StringVar(&s.dbDSN, "d", s.dbDSN, "database dsn")

			fs.Parse(configOptions.osArgs)
		}
	}

	return &s
}

func (c Config) SessionLifetime() time.Duration {
	return c.sessionLifeTime
}

func (c Config) AccrualSysAddr() string {
	return c.accSystem
}

func (c Config) DatabaseDSN() string {
	return c.dbDSN
}

func (c Config) RunAddress() string {
	return c.runAddr
}
