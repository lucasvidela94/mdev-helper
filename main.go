// mdev is a CLI tool for managing mobile development environments.
package main

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
