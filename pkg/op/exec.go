package op

import (
	"os"
	"os/exec"
	"strings"
)

func GetExecPath() (*string, error) {
	opLocation, err := exec.Command("/usr/bin/which", "op").Output()
	if err != nil {
		return nil, err
	}

	opPath := strings.TrimSpace(string(opLocation))
	return &opPath, nil
}

func (config *Config) Exec(args []string) (*exec.Cmd, error) {
	opPath, err := GetExecPath()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(*opPath, args...)

	envVars := config.GetEnvVars()
	cmd.Env = append(os.Environ(), envVars...)
	return cmd, nil
}
