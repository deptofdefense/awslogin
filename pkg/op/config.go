package op

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	SessionName  string `json:"session_name"`
	SessionToken string `json:"session_token"`
	Expiration   int64  `json:"expiration"`
}

func New(sessionName, sessionToken string) *Config {
	config := &Config{
		SessionName:  sessionName,
		SessionToken: sessionToken,
	}
	config.ResetExpiration()
	return config
}

// TODO: Remove
func (config *Config) ResetExpiration() {
	config.Expiration = time.Now().AddDate(0, 0, 30).Unix()
}

func (config *Config) GetEnvVars() []string {
	envVars := []string{
		fmt.Sprintf("%s=%s", config.SessionName, config.SessionToken),
		fmt.Sprintf("OP_EXPIRATION=%d", config.Expiration),
	}
	return envVars
}

func CheckSession(sessionFilename string) (*Config, error) {
	if _, err := os.Stat(sessionFilename); os.IsNotExist(err) {
		return nil, fmt.Errorf("The session filename does not exist at %s\n", sessionFilename)
	}

	config, err := LoadConfig(sessionFilename)
	if err != nil {
		return nil, err
	}

	_, err = config.GetAccount()
	if err != nil {
		config, err = Signin(sessionFilename)
		if err != nil {
			return nil, err
		}
		_, err = config.GetAccount()
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	return config, nil
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func WriteConfig(filename string, config *Config) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	data, errMarshal := json.Marshal(config)
	if errMarshal != nil {
		return errMarshal
	}
	_, errWrite := f.Write(data)
	if errWrite != nil {
		return errWrite
	}
	return nil
}
