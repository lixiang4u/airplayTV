package model

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
	"log"
)

var orm gorose.IOrm

// mysql示例, 记得导入mysql驱动 github.com/go-sql-driver/mysql
func Connect() gorose.IOrm {
	// dsn: "root:root@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=true"
	var dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.db"),
		viper.GetString("mysql.charset"),
	)
	engin, err := gorose.Open(&gorose.Config{
		Driver: "mysql",
		Dsn:    dsn,
	})
	if err != nil {
		log.Fatalln(err)
	}
	orm = engin.NewOrm()
}

func NewEngin() gorose.IOrm {
	if orm == nil {
		orm = Connect()
	}
	return orm
}
