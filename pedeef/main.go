package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	commands2 "pedeef/commands"
)

var cmd = initCommand()

const (
	UtilName = "pedeef"
)

func initCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: UtilName,
		Run: func(cmd *cobra.Command, a_ []string) {
			fmt.Println("Available commands:")
			for _, command := range cmd.Commands() {
				fmt.Printf("  %s: %s\n", command.Name(), command.Short)
			}
		},
	}

	cmd.AddCommand(commands2.NewMergeCommand())
	cmd.AddCommand(commands2.NewCompressCommand())
	return cmd
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Process failed with an error: [%s]\n", err)
		os.Exit(1)
	}
}
