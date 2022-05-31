package main

import (
	"github.com/ichaly/go-env"
	_ "github.com/ichaly/go-env/auto"
	"log"
	"os"
)

type Redis struct {
	Password string `env:"PASSWORD,default=redis123"`
}
type Mysql struct {
	Password string `env:"PASSWORD,default=mysql123"`
}
type Config struct {
	Port     int     `env:"PORT,default=8080"`
	Username string  `env:"USERNAME,required=true"`
	Cache    bool    `env:"CACHE1"`
	Price    float32 `env:"PRICE,default=0.0"`
	Redis    *Redis
	Mysql    Mysql
	Keys     []float64
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
	log.Printf("%v*****>>>%v", cfg, cfg.Redis.Password)
}
