package main

import "log"

func main() {
	server, err := NewServer("./database.db")
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start()
	if err != nil {
		print(err)
	}

}
