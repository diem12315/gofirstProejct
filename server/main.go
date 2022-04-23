package main

import "example.com/m/TelPackage"

func main() {
	server := TelPackage.NewServer("127.0.0.1", 8888)
	server.Start()
}
