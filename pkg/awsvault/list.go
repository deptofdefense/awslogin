package awsvault

import (
	"time"

	"github.com/99designs/aws-vault/v6/vault"
	"github.com/99designs/keyring"
)

func GetProfiles(f *vault.ConfigFile) ([]string, error) {

	return f.ProfileNames(), nil
}

func GetSessions(f *vault.ConfigFile, keyring keyring.Keyring) (map[string]time.Duration, error) {
	profileSessions := map[string]time.Duration{}

	credentialKeyring := &vault.CredentialKeyring{Keyring: keyring}
	sessionKeyring := &vault.SessionKeyring{Keyring: credentialKeyring.Keyring}

	sessions, err := sessionKeyring.GetAllMetadata()
	if err != nil {
		return profileSessions, err
	}

	for _, profileName := range f.ProfileNames() {

		// check session keyring
		for _, sess := range sessions {
			if profileName == sess.ProfileName {
				profileSessions[profileName] = time.Until(sess.Expiration)
			}
		}
	}

	return profileSessions, nil
}
