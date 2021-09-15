package main

import (
	"bufio"
	"errors"
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
	flagLoginBrowser     = "browser"
	flagLoginSectionName = "section-name"
	flagLoginFieldTitle  = "field-title"

	flagSessionDirectory = "session-directory"
	flagSessionFilename  = "session-filename"

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
	flag.String(flagLoginBrowser, browserChrome, "The browser to open with")
	flag.String(flagLoginSectionName, "ACCOUNT_INFO", "The 1Password section name used to identify AWS credentials")
	flag.String(flagLoginFieldTitle, "ACCOUNT_ALIAS", "The 1Password field title used to identify AWS Account Alias")
	flag.String(flagSessionDirectory, HOMEDIR, "The path of the directory to hold the session information")
	flag.String(flagSessionFilename, SESSION_FILE, "The name of the file to retain session information")
}

func checkLoginConfig(v *viper.Viper) error {
	browser := v.GetString(flagLoginBrowser)
	if _, ok := browserToPath[browser]; !ok {
		return fmt.Errorf("Given browser %q is not an option\n", browser)
	}
	sessionDirectory := v.GetString(flagSessionDirectory)
	if sessionDirectory == HOMEDIR {
		homedir, errUserHomeDir := os.UserHomeDir()
		if errUserHomeDir != nil {
			return errUserHomeDir
		}
		sessionDirectory = homedir
	}
	if _, err := os.Stat(sessionDirectory); os.IsNotExist(err) {
		return fmt.Errorf("The session directory %q does not exist\n", sessionDirectory)
	}
	sessionFilename := v.GetString(flagSessionFilename)
	if len(sessionFilename) == 0 {
		return errors.New("The session filename should not be empty")
	}
	return nil
}

func login(cmd *cobra.Command, args []string) error {
	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w\n", errViper)
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
	sectionName := v.GetString(flagLoginSectionName)
	fieldTitle := v.GetString(flagLoginFieldTitle)
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
	// TODO: Sort by name of title
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
		return fmt.Errorf("No entries were found using filters %v\n", filters)
	}

	item, errGetItem := config.GetItem(title)
	if errGetItem != nil {
		return errGetItem
	}

	var accountAlias string
	for _, section := range item.Details.Sections {
		if section.Title == sectionName {
			for _, field := range section.Fields {
				if field.T == fieldTitle {
					accountAlias = field.V
				}
			}
		}
	}

	if len(strings.TrimSpace(accountAlias)) == 0 {
		return fmt.Errorf("There is no account alias defined for the choice %q\n", title)
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
