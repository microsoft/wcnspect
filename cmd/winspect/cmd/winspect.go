package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	winspectCmd := newCommandsBuilder().addAll().build()
	cmd := winspectCmd.getCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
