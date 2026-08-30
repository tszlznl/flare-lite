package main

import (
	"nav/cmd"
	"nav/internal/server"
)

func main() {
	flags := cmd.Parse()
	server.StartDaemon(&flags)
}
