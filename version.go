package main

import (
	"fmt"

	"github.com/deptofdefense/awslogin/pkg/version"

	"github.com/spf13/cobra"
)

func printVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("%s version %s\n", CLI_NAME, version.Full())
	return nil
}
