package op

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/99designs/aws-vault/v6/prompt"
)

func Signin(sessionFilename string) (*Config, error) {

	opPath, errGetExecPath := GetExecPath()
	if errGetExecPath != nil {
		return nil, errGetExecPath
	}

	pass, errTerminalSecretPrompt := prompt.TerminalSecretPrompt("Enter your 1Password password: ")
	if errTerminalSecretPrompt != nil {
		return nil, errTerminalSecretPrompt
	}

	command := exec.Command(*opPath, "signin")
	stdin, errStdinPipe := command.StdinPipe()
	if errStdinPipe != nil {
		log.Fatal(errStdinPipe)
	}

	go func() {
		defer stdin.Close()
		_, errWriteString := io.WriteString(stdin, pass)
		if errWriteString != nil {
			log.Fatal(errWriteString)
		}
	}()

	out, errOutput := command.Output()
	if errOutput != nil {
		log.Fatal(errOutput)
	}
	// The output appears to be:
	// export OP_SESSION_dds="_N8UtA6Y-NGyiWycztN9PZbuDA0g-B7xXOkrIGD1E91"
	opSession := strings.Split(strings.Split(strings.Split(string(out), "\n")[0], " ")[1], "=")
	sessionName := opSession[0]
	sessionToken := strings.Trim(opSession[1], "\"")

	config := New(sessionName, sessionToken)
	errWriteConfig := WriteConfig(sessionFilename, config)
	if errWriteConfig != nil {
		return nil, errWriteConfig
	}

	fmt.Printf("1Password session file saved to: %s\n", sessionFilename)
	return config, nil
}
