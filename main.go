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

	//	var v = url.Values{}
	//	v.Add("url","id")
	//	v.Add("sign",string(time.Now().Unix()))
	//	log.Println("[en]",v.Encode())
	//
	//os.Exit(-1)
	//	controller.GetNNVideoUrl("101451")

	cmd.Execute()
	// url=iXCsMUeAcFkKpylSQk0i8ojsZ7usfjlV%2BVBxdO3cOeaUoaa%2F90XoX9dfjkAC32NkxfUiad1p%2BLPirkYRDvf2AZK7nJeJBh8WorlWDvRVkac%3D&sign=1655391285
}
