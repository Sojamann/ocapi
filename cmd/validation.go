package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// returns error if string is invalid else nil
type argValidator func(string) error

func validateArgNo(num int, fn argValidator) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < num {
			panic(fmt.Sprintf("Cannot validate argument at index %d, only %d were given", num, len(args)))
		}

		return fn(args[num])
	}
}
