package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CLI_NAME     = "awslogin"
	HOMEDIR      = "~/"
	SESSION_FILE = ".op_session"
)

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()
	errBind := v.BindPFlags(cmd.Flags())
	if errBind != nil {
		return v, fmt.Errorf("error binding flag set to viper: %w\n", errBind)
	}
	v.SetEnvPrefix(CLI_NAME) // Enforces all env vars to require "AWSLOGIN_", making them unique
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	return v, nil
}

func main() {
	rootCommand := &cobra.Command{
		Use:                   fmt.Sprintf("%s [flags]", CLI_NAME),
		DisableFlagsInUseLine: true,
		Short:                 "Log into AWS using credentials stored in 1Password",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  login,
	}
	initLoginFlags(rootCommand.Flags())

	if err := rootCommand.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", CLI_NAME, err.Error())
		_, _ = fmt.Fprintf(os.Stderr, "Try %s --help for more information.\n", CLI_NAME)
		os.Exit(1)
	}
}
