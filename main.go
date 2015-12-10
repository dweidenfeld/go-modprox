package main

import (
	"./proxy"
	"net/http"
	"fmt"
	"log"
)

func main() {
	c, err := proxy.ReadConfigFromFile("./config.json")
	if nil != err {
		panic(err)
	}
	p := proxy.New(c)

	log.Printf("Listening on :%d\n", c.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), p.Handler()); nil != err {
		log.Fatal(err)
	}
}
