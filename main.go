package main

import (
	"github.com/lixiang4u/airplayTV/cmd"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/spf13/viper"
	"log"
)

func init() {
	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	util.InitLog()

	cmd.Execute()

}
