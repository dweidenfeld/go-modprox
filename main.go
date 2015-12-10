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


	if 0 < len(c.SSL.Cert) && 0 < len(c.SSL.Key) {
		log.Printf("Listening on :%d (with SSL)\n", c.Port)
		err = http.ListenAndServeTLS(fmt.Sprintf(":%d", c.Port), c.SSL.Cert, c.SSL.Key, p.Handler())
	} else {
		log.Printf("Listening on :%d\n", c.Port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", c.Port), p.Handler())
	}
	if nil != err {
		log.Fatal(err)
	}
}
