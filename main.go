package main

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// func main() {
// 	dsn := "cp_65011212157:65011212157@csmsu@tcp(202.28.34.197:3306)/cp_65011212157"
// 	dialactor := mysql.Open(dsn)
// 	_, err := gorm.Open(dialactor)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Connection successful")
// }

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(viper.Get("mysql.dsn"))
	dsn := viper.GetString("mysql.dsn")

	dialactor := mysql.Open(dsn)
	_, err = gorm.Open(dialactor)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection successful")
}
