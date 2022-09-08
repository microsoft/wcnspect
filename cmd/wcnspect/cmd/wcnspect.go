// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	wcnspectCmd := newCommandsBuilder().addAll().build()
	cmd := wcnspectCmd.getCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
