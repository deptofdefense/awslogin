package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/deptofdefense/awslogin/pkg/op"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initOPSigninFlags(flag *pflag.FlagSet) {
}

func checkOPSigninConfig(v *viper.Viper) error {
	return nil
}

func opSignin(cmd *cobra.Command, args []string) error {
	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w", errViper)
	}

	if errConfig := checkOPSigninConfig(v); errConfig != nil {
		return errConfig
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	sessionFilename := path.Join(homedir, ".op_session")

	config, err := op.Signin(sessionFilename)
	if err != nil {
		return err
	}

	// Verify that the session token works by getting account info
	_, err = config.GetAccount()
	if err != nil {
		return err
	}

	return nil
}
