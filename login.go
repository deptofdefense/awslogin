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

const (
	flagLoginBrowser = "browser"

	browserChrome          = "chrome"
	browserChromeIncognito = "chrome-incognito"
	browserChromeCanary    = "chrome-canary"
	browserSafari          = "safari"
	browserFirefox         = "firefox"
)

var (
	browserPathChrome          = []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "--new-window"}
	browserPathChromeIncognito = []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "--new-window", "--args", "--incognito"}
	browserPathChromeCanary    = []string{"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary", "--new-window"}
	browserPathSafari          = []string{"/usr/bin/open", "-a", "/Applications/Safari.app/Contents/MacOS/Safari"}
	browserPathFirefox         = []string{"/Applications/Firefox.app/Contents/MacOS/firefox"}
	browserToPath              = map[string][]string{
		browserChrome:          browserPathChrome,
		browserChromeIncognito: browserPathChromeIncognito,
		browserChromeCanary:    browserPathChromeCanary,
		browserSafari:          browserPathSafari,
		browserFirefox:         browserPathFirefox,
	}
)

func initLoginFlags(flag *pflag.FlagSet) {
	flag.String(flagLoginBrowser, browserFirefox, "The browser to open with")
	flag.String(flagSessionDirectory, HOMEDIR, "The path of the directory to hold the session information")
	flag.String(flagSessionFilename, ".op_session", "The name of the file to retain session information")
}

func checkLoginConfig(v *viper.Viper) error {
	browser := v.GetString(flagLoginBrowser)
	if _, ok := browserToPath[browser]; !ok {
		return fmt.Errorf("Given browser %q is not an option", browser)
	}
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

	browser := v.GetString(flagLoginBrowser)
	browserPath := browserToPath[browser]
	sessionDirectory := v.GetString(flagSessionDirectory)
	sessionFilename := v.GetString(flagSessionFilename)

	if sessionDirectory == HOMEDIR {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		sessionDirectory = homedir
	}
	sessionPath := path.Join(sessionDirectory, sessionFilename)

	config, err := op.CheckSession(sessionPath)
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

	if len(strings.TrimSpace(accountAlias)) == 0 {
		return fmt.Errorf("There is no account alias defined for the choice %d %q", numChoice, title)
	}

	fmt.Printf("Account Alias: %s\n", accountAlias)

	totp, err := config.GetTotp(title)
	if err != nil {
		return err
	}

	oneTimePassword := strings.TrimSpace(*totp)
	fmt.Printf("MFA Token: %s\n", oneTimePassword)

	// Create the commands to use
	command1 := exec.Command("/usr/local/bin/aws-vault", "login", accountAlias, "--mfa-token", oneTimePassword, "--stdout")
	fmt.Println(command1.String())
	command2 := exec.Command("xargs", append([]string{"-t"}, browserPath...)...)
	fmt.Println(command2.String())

	// Set up the pipe
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return err
	}
	command1.Stdout = writePipe
	command2.Stdin = readPipe
	command2.Stdout = os.Stdout
	command1.Start()
	command2.Start()

	go func() {
		defer writePipe.Close()
		command1.Wait()
	}()
	_ = command2.Run()

	return nil
}
