package main

import (
	_ "github.com/jackc/pgx/v5"

	"github.com/vhakulinen/dino/cli"
)

func main() {
	cmd, _ := cli.RootCommand("dino-dev", "pgx")
	cmd.Execute()
}
