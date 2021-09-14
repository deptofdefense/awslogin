package op

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func Signin(sessionFilename string) (*Config, error) {

	opPath, err := GetExecPath()
	if err != nil {
		return nil, err
	}

	fmt.Println("Enter your 1Password password:")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatal(err)
	}
	pass := strings.TrimSpace(string(bytePassword))

	command := exec.Command(*opPath, "signin")
	stdin, err := command.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, pass)
	}()

	out, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}
	// The output appears to be:
	// export OP_SESSION_dds="_N8UtA6Y-NGyiWycztN9PZbuDA0g-B7xXOkrIGD1E91"
	opSession := strings.Split(strings.Split(strings.Split(string(out), "\n")[0], " ")[1], "=")
	sessionName := opSession[0]
	sessionToken := strings.Trim(opSession[1], "\"")

	config := New(sessionName, sessionToken)
	WriteConfig(sessionFilename, config)

	fmt.Printf("1Password session file saved to: %s\n", sessionFilename)
	return config, nil
}
