package main

import (
	_ "github.com/lib/pq"

	"github.com/vhakulinen/dino/commands"
)

func main() {
	cmd := commands.RootCommand("dino-dev")
	cmd.Execute()
}
