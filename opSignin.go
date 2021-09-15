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
	initSessionFlags(flag)
}

func checkOPSigninConfig(v *viper.Viper) error {
	errCheckSessionConfig := checkSessionConfig(v)
	if errCheckSessionConfig != nil {
		return errCheckSessionConfig
	}
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

	sessionDirectory := v.GetString(flagSessionDirectory)
	sessionFilename := v.GetString(flagSessionFilename)

	if sessionDirectory == HOMEDIR {
		homedir, errUserHomeDir := os.UserHomeDir()
		if errUserHomeDir != nil {
			log.Fatal(errUserHomeDir)
		}
		sessionDirectory = homedir
	}
	sessionPath := path.Join(sessionDirectory, sessionFilename)

	_, err := op.CheckSession(sessionPath)
	if err != nil {
		return err
	}

	return nil
}
