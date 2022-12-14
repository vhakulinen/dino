package main

import (
	_ "github.com/jackc/pgx/v5"

	"github.com/vhakulinen/dino/cli"
)

func main() {
	cmd, _ := cli.New(cli.OptionCmdName("dino-dev"), cli.OptionDbDriver("pgx"))
	cmd.Execute()
}
