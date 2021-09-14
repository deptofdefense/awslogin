package op

import (
	"encoding/json"
)

func (config *Config) GetItem(name string) (*Item, error) {
	var item Item

	command, err := config.Exec([]string{"get", "item", name})
	if err != nil {
		return &item, err
	}
	out, err := command.Output()
	if err != nil {
		return &item, err
	}
	json.Unmarshal(out, &item)

	return &item, nil
}

func (config *Config) ListItems(tags string) ([]Item, error) {
	// 1p list items --tags $1 --categories login | jq -Mcr '.[].overview.title' | sort
	var items []Item

	command, err := config.Exec([]string{"list", "items", "--tags", tags, "--categories", "login"})
	if err != nil {
		return items, err
	}
	out, err := command.Output()
	if err != nil {
		return items, err
	}
	json.Unmarshal(out, &items)

	return items, nil
}

func (config *Config) GetAccount() (*string, error) {
	command, err := config.Exec([]string{"get", "account"})
	if err != nil {
		return nil, err
	}
	out, err := command.Output()
	if err != nil {
		return nil, err
	}
	strOut := string(out)
	return &strOut, nil
}

func (config *Config) GetTotp(name string) (*string, error) {

	command, err := config.Exec([]string{"get", "totp", name})
	if err != nil {
		return nil, err
	}
	out, err := command.Output()
	if err != nil {
		return nil, err
	}
	strOut := string(out)
	return &strOut, nil
}
