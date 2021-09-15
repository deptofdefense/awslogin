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

	var filters []string
	if len(args) > 0 {
		filters = args
	}

	browser := v.GetString(flagLoginBrowser)
	browserPath := browserToPath[browser]
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

	config, errCheckSession := op.CheckSession(sessionPath)
	if errCheckSession != nil {
		return errCheckSession
	}

	tags := "aws"
	items, errListItems := config.ListItems(tags)
	if errListItems != nil {
		return errListItems
	}

	// Filter the items first
	newItemList := []op.Item{}
	if len(filters) > 0 {
		for _, item := range items {
			for _, f := range filters {
				title := item.Overview.Title
				if strings.Contains(title, f) {
					newItemList = append(newItemList, item)
				}
			}
		}
	} else {
		newItemList = items
	}

	var title string
	if len(newItemList) > 1 {
		for num, item := range newItemList {
			fmt.Println(num, item.Overview.Title)
		}

		fmt.Printf("\nChoose a secret's number: ")
		reader := bufio.NewReader(os.Stdin)
		choice, errReadString := reader.ReadString('\n')
		if errReadString != nil {
			return errReadString
		}
		numChoice, errAtoi := strconv.Atoi(strings.TrimSpace(choice))
		if errAtoi != nil {
			return errAtoi
		}
		title = newItemList[numChoice].Overview.Title
		fmt.Printf("\nYou chose: %s\n\n", title)
	} else if len(newItemList) == 1 {
		title = newItemList[0].Overview.Title
	} else {
		return fmt.Errorf("No entries were found using filters %v", filters)
	}

	item, errGetItem := config.GetItem(title)
	if errGetItem != nil {
		return errGetItem
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
		return fmt.Errorf("There is no account alias defined for the choice %q", title)
	}

	fmt.Printf("Account Alias: %s\n", accountAlias)

	totp, errGetTotp := config.GetTotp(title)
	if errGetTotp != nil {
		return errGetTotp
	}

	oneTimePassword := strings.TrimSpace(*totp)
	fmt.Printf("MFA Token: %s\n", oneTimePassword)

	// Create the commands to use
	command1 := exec.Command("/usr/local/bin/aws-vault", "login", accountAlias, "--mfa-token", oneTimePassword, "--stdout")
	command2 := exec.Command("xargs", append([]string{"-t"}, browserPath...)...)

	// Set up the pipe
	readPipe, writePipe, errPipe := os.Pipe()
	if errPipe != nil {
		return errPipe
	}
	command1.Stdout = writePipe
	command2.Stdin = readPipe
	command2.Stdout = os.Stdout
	errStart1 := command1.Start()
	if errStart1 != nil {
		return errStart1
	}
	errStart2 := command2.Start()
	if errStart2 != nil {
		return errStart2
	}

	go func() {
		defer writePipe.Close()
		_ = command1.Wait()
	}()
	_ = command2.Run()

	return nil
}
