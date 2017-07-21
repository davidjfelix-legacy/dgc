package cmd

import (
	"github.com/spf13/cobra"
	"time"
)

var RootCmd = &cobra.Command{
	Use:   "dgc",
	Short: "dgc is a docker garbage collector",
}

var all bool
var force bool
var grace time.Duration

func init() {
	RootCmd.PersistentFlags().
		BoolVarP(&all, "all", "a", false, "Ignore grace TTL timeout and clean everything")
	RootCmd.PersistentFlags().
		BoolVarP(&force, "force", "f", false, "Forcibly remove any resources slated to be cleaned")
	RootCmd.PersistentFlags().
		DurationVarP(&grace, "grace", "g", 1*time.Hour, "The TTL grace timeout on docker resources to be cleaned")
}
