package main

import (
	_ "github.com/ichaly/go-env/autoload"
	"log"
	"os"
)

func main() {
	log.Printf("+++++>>>%v", os.Getenv("TEST"))
}
