package main

import (
	"github.com/lixiang4u/airplayTV/cmd"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/spf13/viper"
	"log"
)

func init() {
	util.LoadConfig()
}

func main() {
	util.InitLog()

	cmd.Execute()

}
