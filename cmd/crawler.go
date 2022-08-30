package cmd

import (
	"fmt"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/spf13/cobra"
	"log"
)

var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "start crawler jobs(not implemented!!!)",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(fmt.Sprintf("[AppPath] %s", util.AppPath()))
	},
}

func init() {
	rootCmd.AddCommand(crawlerCmd)
}
