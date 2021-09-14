package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/deptofdefense/awslogin/pkg/op"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initLoginFlags(flag *pflag.FlagSet) {
}

func checkLoginConfig(v *viper.Viper) error {
	return nil
}

func login(cmd *cobra.Command, args []string) error {
	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w", errViper)
	}

	if errConfig := checkLoginConfig(v); errConfig != nil {
		return errConfig
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	sessionFilename := path.Join(homedir, ".op_session")

	config, err := op.CheckSession(sessionFilename)
	if err != nil {
		return err
	}

	tags := "aws"
	items, err := config.ListItems(tags)
	if err != nil {
		return err
	}
	for num, item := range items {
		fmt.Println(num, item.Overview.Title)
	}

	fmt.Printf("\nChoose a secret's number: ")
	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	numChoice, err := strconv.Atoi(strings.TrimSpace(choice))
	if err != nil {
		return err
	}
	title := items[numChoice].Overview.Title
	fmt.Printf("\nYou chose: %s\n\n", title)

	item, err := config.GetItem(title)
	if err != nil {
		return err
	}

	var accountAlias string
	for _, section := range item.Details.Sections {
		if section.Title == "ACCOUNT_INFO" {
			for _, field := range section.Fields {
				if field.T == "Account Alias" {
					accountAlias = field.V
				}
			}
		}
	}

	fmt.Printf("Account Alias: %s\n", accountAlias)

	totp, err := config.GetTotp(title)
	if err != nil {
		return err
	}

	oneTimePassword := strings.TrimSpace(*totp)
	fmt.Printf("MFA Token: %s\n", oneTimePassword)

	command := exec.Command("/usr/local/bin/aws-vault", "login", accountAlias, "--mfa-token", oneTimePassword)
	out, err := command.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}
