package main

import (
	"log"
)

func main() {

	mongostore, err := ConnectToMongoDB("mongodb://127.0.0.1:27017", "sample", "sample")
	if err != nil {
		log.Fatal(err)
	}


	// postgres, err := NewPostgresStore("user=postgres dbname=postgres password=user sslmode=disable")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := postgres.Init(); err != nil {
	// 	log.Fatal(err)
	// }

	server := NewAPIServer(":3000", mongostore)
	server.Run()
}
