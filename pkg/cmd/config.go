package cmd

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/cobra"
)

var Config = ConfigStruct{
	DatabaseURI:          "",
	RunAddress:           "localhost:8080",
	AccrualSystemAddress: "",
	LogLevel:             "info",
}

type ConfigStruct struct {
	DatabaseURI          string `env:"DATABASE_URI"`
	RunAddress           string `env:"RUN_ADDRESS"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
}

func (c *ConfigStruct) initEnvArgs() error {
	if err := env.Parse(c); err != nil {
		return err
	}
	return nil
}

func (c *ConfigStruct) validate() error {
	if c.DatabaseURI == "" {
		return fmt.Errorf("database uri can not be empty")
	}
	if c.RunAddress == "" {
		return fmt.Errorf("Run address uri can not be empty")
	}

	if c.AccrualSystemAddress == "" {
		return fmt.Errorf("accrual address uri can not be empty")
	}
	return nil
}

func setupConfig(cmd *cobra.Command, args []string) error {
	cmd.DisableFlagParsing = false

	cmd.Flags().StringVarP(&Config.DatabaseURI, "database_uri", "d", Config.DatabaseURI, "Postgre URI")
	cmd.Flags().StringVarP(&Config.RunAddress, "run_addr", "a", Config.RunAddress, "Run address")
	cmd.Flags().StringVarP(&Config.AccrualSystemAddress, "accrual_addr", "r", Config.AccrualSystemAddress, "Accrual system address")
	cmd.Flags().StringVarP(&Config.LogLevel, "log_level", "l", Config.LogLevel, "Log level")

	if err := cmd.ParseFlags(args); err != nil {
		return err
	}

	if err := Config.initEnvArgs(); err != nil {
		return err
	}
	if err := Config.validate(); err != nil {
		return err
	}
	return nil
}
