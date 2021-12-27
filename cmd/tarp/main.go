package main

import (
	"github.com/dylandreimerink/tarp"
	"github.com/spf13/cobra"
)

func main() {
	makeCmd().Execute()
}

var flagOutput string

func makeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "tarp {coverage file 1} [coverage file N...]",
		Long: "Tarp generates an interactive HTML coverage report for Go coverage files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tarp.GenerateHTMLReport(args, flagOutput)
		},
		Args: cobra.MinimumNArgs(1),
	}

	f := cmd.Flags()
	f.StringVarP(&flagOutput, "output", "o", "./coverage.html", "The generated coverage report")

	return cmd
}
