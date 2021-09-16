package awsvault

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/99designs/aws-vault/v6/cli"
	"github.com/99designs/aws-vault/v6/vault"
)

func GetProfiles() ([]string, error) {
	// Disable the logging from the vault package
	log.SetOutput(ioutil.Discard)

	awsVault := &cli.AwsVault{}

	awsConfigFile, err := awsVault.AwsConfigFile()
	if err != nil {
		return []string{}, err
	}

	return awsConfigFile.ProfileNames(), nil
}

func GetSessions() (map[string]time.Duration, error) {
	// Disable the logging from the vault package
	log.SetOutput(ioutil.Discard)

	profileSessions := map[string]time.Duration{}

	awsVault := &cli.AwsVault{}
	keyring, err := awsVault.Keyring()
	if err != nil {
		return profileSessions, err
	}

	awsConfigFile, err := awsVault.AwsConfigFile()
	if err != nil {
		return profileSessions, err
	}

	credentialKeyring := &vault.CredentialKeyring{Keyring: keyring}
	sessionKeyring := &vault.SessionKeyring{Keyring: credentialKeyring.Keyring}

	sessions, err := sessionKeyring.GetAllMetadata()
	if err != nil {
		return profileSessions, err
	}

	for _, profileName := range awsConfigFile.ProfileNames() {

		// check session keyring
		for _, sess := range sessions {
			if profileName == sess.ProfileName {
				profileSessions[profileName] = time.Until(sess.Expiration)
			}
		}
	}

	return profileSessions, nil
}
