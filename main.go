// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package main

import (
	"fmt"
	"os"

	"github.com/bytefreezer/fakedata/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
