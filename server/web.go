package main


func main() {
	server, err := NewServer("./database.db")
	if err != nil {
		print(err)
	}

	err = server.Start()
	if err != nil {
		print(err)
	}

}
