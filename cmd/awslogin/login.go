package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/99designs/aws-vault/v6/cli"
	"github.com/deptofdefense/awslogin/pkg/awsvault"
	"github.com/deptofdefense/awslogin/pkg/op"
	"github.com/deptofdefense/awslogin/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
)

const (
	flagLoginBrowser     = "browser"
	flagLoginSectionName = "section-name"
	flagLoginFieldTitle  = "field-title"
	flagLoginVersion     = "version"
	flagLoginVerbose     = "verbose"

	flagSessionDirectory = "session-directory"
	flagSessionFilename  = "session-filename"

	browserChrome          = "chrome"
	browserChromeIncognito = "chrome-incognito"
	browserChromeCanary    = "chrome-canary"
	browserSafari          = "safari"
	browserFirefox         = "firefox"

	minVersionOP = "1.11.4"
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
	flag.Bool(flagLoginVersion, false, "Display the version information and exit")
	flag.Bool(flagLoginVerbose, false, "Use verbose output")
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

// preCheck will return an error if prerequisites are not met
func preCheck(commandName string, commandArgs []string, expected string) error {

	commandPath, errEvalSymlinks := filepath.EvalSymlinks(commandName)
	if errEvalSymlinks != nil {
		return errEvalSymlinks
	}
	versionCommand := exec.Command(commandPath, commandArgs...)
	actual, errOutput := versionCommand.CombinedOutput()
	if errOutput != nil {
		return fmt.Errorf("Unable to call version command for %q with args %v: %w", commandName, commandArgs, errOutput)
	}
	actualStr := strings.TrimSpace(string(actual))
	if len(actualStr) == 0 {
		return fmt.Errorf("No output returned for version command for %q with args %v", commandName, commandArgs)
	}
	// Prefix with 'v' for comparison sake
	if actualStr[0] != 'v' {
		actualStr = "v" + actualStr
	}
	if expected[0] != 'v' {
		expected = "v" + expected
	}
	// Here we will be using semver.Compare(Actual, Expected)
	if semver.Compare(actualStr, expected) < 0 {
		return fmt.Errorf("Expected version of %q to be greater or equal to %q", commandName, expected)
	}
	return nil
}

func login(cmd *cobra.Command, args []string) error {
	// Disable the logging from the vault package
	log.SetOutput(ioutil.Discard)

	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w\n", errViper)
	}

	if v.GetBool(flagLoginVersion) {
		fmt.Println(version.Full())
		return nil
	}

	if errConfig := checkLoginConfig(v); errConfig != nil {
		return errConfig
	}

	// Confirm that the minimum version is met for these tools
	errPreCheck := preCheck("/usr/local/bin/op", []string{"--version"}, minVersionOP)
	if errPreCheck != nil {
		return errPreCheck
	}

	// AWS_PROFILE is a special env var which can be used to immediately log in
	accountAlias := os.Getenv("AWS_PROFILE")

	// Handle Args
	var filters []string
	if len(args) > 0 {
		// When using filters the AWS_PROFILE should be added to the list
		filters = args
		if len(accountAlias) > 0 {
			filters = append(filters, accountAlias)
			// The accountAlias is set to an empty string to ensure that the filters are used
			accountAlias = ""
		}
	}

	// Handle Flags
	browser := v.GetString(flagLoginBrowser)
	browserPath := browserToPath[browser]
	sectionName := v.GetString(flagLoginSectionName)
	fieldTitle := v.GetString(flagLoginFieldTitle)
	sessionDirectory := v.GetString(flagSessionDirectory)
	sessionFilename := v.GetString(flagSessionFilename)
	verbose := v.GetBool(flagLoginVerbose)

	// Get the session path for using 1Password
	if sessionDirectory == HOMEDIR {
		homedir, errUserHomeDir := os.UserHomeDir()
		if errUserHomeDir != nil {
			return errUserHomeDir
		}
		sessionDirectory = homedir
	}
	sessionPath := path.Join(sessionDirectory, sessionFilename)

	awsVault := &cli.AwsVault{}
	keyring, err := awsVault.Keyring()
	if err != nil {
		return err
	}

	awsConfigFile, err := awsVault.AwsConfigFile()
	if err != nil {
		return err
	}

	// See if an active session exists already
	profileSessions, err := awsvault.GetSessions(awsConfigFile, keyring)
	if err != nil {
		return err
	}

	var loginURL *string
	var errGetLoginURL error
	var title string

	if len(accountAlias) == 0 {
		config, errCheckSession := op.CheckSession(sessionPath)
		if errCheckSession != nil {
			return errCheckSession
		}

		var errChooseAccountAlias error
		title, accountAlias, errChooseAccountAlias = chooseAccountAlias(config, sectionName, fieldTitle, filters)
		if errChooseAccountAlias != nil {
			return errChooseAccountAlias
		}
	}

	sessionDuration, ok := profileSessions[accountAlias]

	// If no active session or the session duration is negative then get the OTP again
	if ok && sessionDuration > 0 {
		loginURL, errGetLoginURL = awsvault.GetLoginURL(accountAlias, "", awsConfigFile, keyring)
		if errGetLoginURL != nil {
			return errGetLoginURL
		}
	} else {
		config, errCheckSession := op.CheckSession(sessionPath)
		if errCheckSession != nil {
			return errCheckSession
		}
		// A safety switch to ensure a title exists
		if len(title) == 0 && len(accountAlias) > 0 {
			title = fmt.Sprintf("AWS %s", accountAlias)
		}
		totp, errGetTotp := config.GetTotp(title)
		if errGetTotp != nil {
			return errGetTotp
		}

		oneTimePassword := strings.TrimSpace(*totp)
		if verbose {
			fmt.Printf("MFA Token: %s\n", oneTimePassword)
		}

		loginURL, errGetLoginURL = awsvault.GetLoginURL(accountAlias, oneTimePassword, awsConfigFile, keyring)
		if errGetLoginURL != nil {
			return errGetLoginURL
		}
	}

	if verbose {
		fmt.Printf("Account Alias: %s\n", accountAlias)
	}

	// Create the commands to use
	command := exec.Command(browserPath[0], append(browserPath[1:], *loginURL)...)

	errStart := command.Start()
	if errStart != nil {
		return errStart
	}

	return nil
}

func chooseAccountAlias(config *op.Config, sectionName, fieldTitle string, filters []string) (string, string, error) {

	tags := "aws"
	items, errListItems := config.ListItems(tags)
	if errListItems != nil {
		return "", "", errListItems
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

	// Sort the items for consistency
	sort.SliceStable(newItemList, func(i, j int) bool {
		return newItemList[i].Overview.Title < newItemList[j].Overview.Title
	})

	var title string
	if len(newItemList) > 1 {
		for num, item := range newItemList {
			fmt.Println(num, item.Overview.Title)
		}

		fmt.Printf("\nChoose the account number: ")
		reader := bufio.NewReader(os.Stdin)
		choice, errReadString := reader.ReadString('\n')
		if errReadString != nil {
			return "", "", errReadString
		}
		numChoice, errAtoi := strconv.Atoi(strings.TrimSpace(choice))
		if errAtoi != nil {
			return "", "", errAtoi
		}
		title = newItemList[numChoice].Overview.Title
		fmt.Printf("\nChosen account: %s\n\n", title)
	} else if len(newItemList) == 1 {
		title = newItemList[0].Overview.Title
	} else {
		return "", "", fmt.Errorf("No entries were found using filters %v\n", filters)
	}

	item, errGetItem := config.GetItem(title)
	if errGetItem != nil {
		return "", "", errGetItem
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
		return "", "", fmt.Errorf("There is no account alias defined for the choice %q\n", title)
	}
	return title, accountAlias, nil
}
