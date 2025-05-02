package main

import (
	"envzilla"
	"fmt"
	"log"
	"os"
)

func main() {
	err := envzilla.Loader("configs/.env")
	if err != nil {
		log.Fatal("WTF : ", err)
	}
	fmt.Println(os.Getenv("kairat"))
}
