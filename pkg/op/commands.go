package op

import (
	"encoding/json"
)

func (config *Config) GetItem(name string) (*Item, error) {
	var item Item

	command, errExec := config.Exec([]string{"get", "item", name})
	if errExec != nil {
		return &item, errExec
	}
	out, errOutput := command.Output()
	if errOutput != nil {
		return &item, errOutput
	}
	errUnmarshal := json.Unmarshal(out, &item)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return &item, nil
}

func (config *Config) ListItems(tags string) ([]Item, error) {
	// 1p list items --tags $1 --categories login | jq -Mcr '.[].overview.title' | sort
	var items []Item

	command, errExec := config.Exec([]string{"list", "items", "--tags", tags, "--categories", "login"})
	if errExec != nil {
		return items, errExec
	}
	out, errOutput := command.Output()
	if errOutput != nil {
		return items, errOutput
	}
	errUnmarshal := json.Unmarshal(out, &items)
	if errUnmarshal != nil {
		return items, errUnmarshal
	}

	return items, nil
}

func (config *Config) GetAccount() (*string, error) {
	command, errExec := config.Exec([]string{"get", "account"})
	if errExec != nil {
		return nil, errExec
	}
	out, errOutput := command.Output()
	if errOutput != nil {
		return nil, errOutput
	}
	strOut := string(out)
	return &strOut, nil
}

func (config *Config) GetTotp(name string) (*string, error) {

	command, errExec := config.Exec([]string{"get", "totp", name})
	if errExec != nil {
		return nil, errExec
	}
	out, errOutput := command.Output()
	if errOutput != nil {
		return nil, errOutput
	}
	strOut := string(out)
	return &strOut, nil
}
