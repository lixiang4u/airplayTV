package main

import (
	"github.com/lixiang4u/ShotTv-api/cmd"
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
	//log.Println("[URL]",service.HandleSrcM3U8FileToLocal("",""))
	//util.InitLog()
	//
	//
	//log.Println("[R]",util.IsUrl("/Users/lixiang4u/repo/ShotTv-api/util/string.go"))
	//log.Println("[R]",util.IsUrl("https://ali2.a.yximgs.com/udata/music/music_541f188013b84118b44ee6f3ff1955fe0.jpg"))
	//log.Println("[R]",util.IsUrl("/udata/music/music_541f188013b84118b44ee6f3ff1955fe0.jpg"))

	cmd.Execute()

}
