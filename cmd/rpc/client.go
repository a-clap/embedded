package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/a-clap/embedded/pkg/embedded"
)

func main() {
	client, err := embedded.NewGPIORPCClient("localhost:50051", time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	
	// Contact the server and print out its response.
	for {
		<-time.After(1 * time.Second)
		r, err := client.Get()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, elem := range r {
			fmt.Println("ID: ", elem.ID)
			fmt.Println("Dir: ", elem.Direction)
			fmt.Println("Active: ", elem.ActiveLevel)
			fmt.Println("Value: ", elem.Value)
		}
		n := r[0]
		n.Value = !n.Value
		
		c, err := client.Configure(n)
		log.Println(c, err)
		
	}
	
}
