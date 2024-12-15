package main

import (
	"fmt"

	"github.com/CZMA-eng/urlshortener/application"
)

func main(){
	a := application.Application{}
	if err := a.Init("./config/config.yaml"); err != nil {
		panic(err)
	}
	fmt.Println("server is starting...")
	a.Run()
	select{}
}