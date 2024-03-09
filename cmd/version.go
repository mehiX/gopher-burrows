package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	version, commit, date = "dev", "dev", "n/a"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print version/build info",
	Long:  "Print version/build information",
	Run: func(cmd *cobra.Command, args []string) {
		printVersion(os.Stdout)
	},
}

func printVersion(out io.Writer) {
	sfmt := "%-20s %s\n"

	fmt.Fprintf(out, sfmt, "Version", version)
	fmt.Fprintf(out, sfmt, "Commit", commit)
	fmt.Fprintf(out, sfmt, "Date", date)
}
