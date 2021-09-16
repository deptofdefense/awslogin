package awsvault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/99designs/aws-vault/v6/vault"
	"github.com/99designs/keyring"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func generateLoginURL(region string, path string) (string, string) {
	loginURLPrefix := "https://signin.aws.amazon.com/federation"
	destination := "https://console.aws.amazon.com/"

	if region != "" {
		destinationDomain := "console.aws.amazon.com"
		switch {
		case strings.HasPrefix(region, "cn-"):
			loginURLPrefix = "https://signin.amazonaws.cn/federation"
			destinationDomain = "console.amazonaws.cn"
		case strings.HasPrefix(region, "us-gov-"):
			loginURLPrefix = "https://signin.amazonaws-us-gov.com/federation"
			destinationDomain = "console.amazonaws-us-gov.com"
		}
		if path != "" {
			destination = fmt.Sprintf("https://%s.%s/%s?region=%s",
				region, destinationDomain, path, region)
		} else {
			destination = fmt.Sprintf("https://%s.%s/console/home?region=%s",
				region, destinationDomain, region)
		}
	}
	return loginURLPrefix, destination
}

func GetLoginURL(profileName string, mfaToken string, f *vault.ConfigFile, keyring keyring.Keyring) (*string, error) {
	vault.UseSession = true

	sessionDuration, errParseDuration := time.ParseDuration("1h")
	if errParseDuration != nil {
		return nil, errParseDuration
	}
	configLoader := vault.ConfigLoader{
		File: f,
		BaseConfig: vault.Config{
			MfaToken:                          mfaToken,
			MfaPromptMethod:                   "terminal",
			NonChainedGetSessionTokenDuration: sessionDuration,
			AssumeRoleDuration:                sessionDuration,
			GetFederationTokenDuration:        sessionDuration,
		},
		ActiveProfile: profileName,
	}
	config, err := configLoader.LoadFromProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("Error loading config: %w", err)
	}

	var creds *credentials.Credentials

	ckr := &vault.CredentialKeyring{Keyring: keyring}
	// If AssumeRole or sso.GetRoleCredentials isn't used, GetFederationToken has to be used for IAM credentials
	if config.HasRole() || config.HasSSOStartURL() {
		creds, err = vault.NewTempCredentials(config, ckr)
	} else {
		creds, err = vault.NewFederationTokenCredentials(profileName, ckr, config)
	}
	if err != nil {
		return nil, fmt.Errorf("profile %s: %w", profileName, err)
	}

	val, err := creds.Get()
	if err != nil {
		return nil, fmt.Errorf("Failed to get credentials for %s: %w", config.ProfileName, err)
	}

	jsonBytes, err := json.Marshal(map[string]string{
		"sessionId":    val.AccessKeyID,
		"sessionKey":   val.SecretAccessKey,
		"sessionToken": val.SessionToken,
	})
	if err != nil {
		return nil, err
	}

	loginURLPrefix, destination := generateLoginURL(config.Region, "")

	req, err := http.NewRequest("GET", loginURLPrefix, nil)
	if err != nil {
		return nil, err
	}

	if expiration, errExpiresAt := creds.ExpiresAt(); errExpiresAt != nil {
		log.Printf("Creating login token, expires in %s", time.Until(expiration))
	}

	q := req.URL.Query()
	q.Add("Action", "getSigninToken")
	q.Add("Session", string(jsonBytes))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Response body was %s", body)
		return nil, fmt.Errorf("Call to getSigninToken failed with %v", resp.Status)
	}

	var respParsed map[string]string

	err = json.Unmarshal([]byte(body), &respParsed)
	if err != nil {
		return nil, err
	}

	signinToken, ok := respParsed["SigninToken"]
	if !ok {
		return nil, fmt.Errorf("Expected a response with SigninToken")
	}

	loginURL := fmt.Sprintf("%s?Action=login&Issuer=aws-vault&Destination=%s&SigninToken=%s",
		loginURLPrefix, url.QueryEscape(destination), url.QueryEscape(signinToken))

	return &loginURL, nil
}
