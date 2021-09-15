package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CLI_NAME = "awslogin"
	HOMEDIR  = "~/"

	flagSessionDirectory = "session-directory"
	flagSessionFilename  = "session-filename"
)

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()
	errBind := v.BindPFlags(cmd.Flags())
	if errBind != nil {
		return v, fmt.Errorf("error binding flag set to viper: %w", errBind)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	return v, nil
}

func main() {
	rootCommand := &cobra.Command{
		Use:                   fmt.Sprintf("%s [flags]", CLI_NAME),
		DisableFlagsInUseLine: true,
		Short:                 "Log into AWS using credentials stored in 1Password",
	}

	loginCommand := &cobra.Command{
		Use:                   `login [flags]`,
		DisableFlagsInUseLine: true,
		Short:                 "login to AWS",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  login,
	}
	initLoginFlags(loginCommand.Flags())

	opSigninCommand := &cobra.Command{
		Use:                   `op-signin [flags]`,
		DisableFlagsInUseLine: true,
		Short:                 "signin to 1Password",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  opSignin,
	}
	initOPSigninFlags(opSigninCommand.Flags())

	versionCommand := &cobra.Command{
		Use:                   `version`,
		DisableFlagsInUseLine: true,
		Short:                 "gitlab POC on events",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  printVersion,
	}

	rootCommand.AddCommand(
		loginCommand,
		opSigninCommand,
		versionCommand,
	)

	if err := rootCommand.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s", CLI_NAME, err.Error())
		_, _ = fmt.Fprintf(os.Stderr, "Try %s --help for more information.", CLI_NAME)
		os.Exit(1)
	}
}
