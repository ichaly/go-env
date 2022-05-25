package main

import (
	"github.com/ichaly/go-env"
	_ "github.com/ichaly/go-env/autoload"
	"log"
	"os"
)

func main() {
	log.Printf("+++++>>>%v", os.Getenv("TEST"))
	res, _ := env.String("Server ip is :${ip:=127.0.0.1},port is ${port:=8080},Hello ${test} !")
	log.Printf("----->>>%v", res)
}
