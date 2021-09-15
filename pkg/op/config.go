package op

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	SessionName  string `json:"session_name"`
	SessionToken string `json:"session_token"`
}

func New(sessionName, sessionToken string) *Config {
	config := &Config{
		SessionName:  sessionName,
		SessionToken: sessionToken,
	}
	return config
}

func (config *Config) GetEnvVars() []string {
	envVars := []string{
		fmt.Sprintf("%s=%s", config.SessionName, config.SessionToken),
	}
	return envVars
}

func CheckSession(sessionFilename string) (*Config, error) {
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
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var config Config
	if len(data) == 0 {
		return &config, nil
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("Unable to read JSON file at %q: %w", filename, err)
	}
	return &config, nil
}

func WriteConfig(filename string, config *Config) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
