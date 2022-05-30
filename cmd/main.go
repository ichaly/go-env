package main

import (
	"github.com/ichaly/go-env"
	_ "github.com/ichaly/go-env/auto"
	"log"
	"os"
)

type Config struct {
	Port     int    `env:"PORT,default=8080"`
	Username string `env:"USERNAME,required=true"`
}

func main() {
	log.Printf("+++++>>>%v", os.Getenv("TEST"))
	res, _ := env.String("Server ip is :${ip:=127.0.0.1},port is ${port:=8080},Hello ${test} !")
	log.Printf("----->>>%v", res)

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
		return
	}
	log.Printf("*****>>>%v", cfg)
}
