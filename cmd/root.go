package cmd

import "github.com/spf13/cobra"

var cmdRoot = &cobra.Command{
	Use:   "burrows",
	Short: "Manage gopher borrows",
	Long:  "Provide an API to assign gophers to burrows and regularly check their status",
}

func Execute() error {
	cmdRoot.AddCommand(cmdServe)
	return cmdRoot.Execute()
}
