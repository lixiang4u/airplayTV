package cmd

import (
	"fmt"
	"github.com/lixiang4u/ShotTv-api/util"
	"github.com/spf13/cobra"
	"log"
)

var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "start crawler jobs",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(fmt.Sprintf("[AppPath] %s", util.AppPath()))
	},
}

func init() {
	rootCmd.AddCommand(crawlerCmd)
}
