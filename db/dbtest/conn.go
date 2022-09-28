package dbtest

import "github.com/vhakulinen/dino/db/utils"

// TODO(ville): Read these variables from env or something.
// TODO(ville): Move this to some internal package so its not exported.
var DefaultConnectionParams = utils.ConnectionParams{
	Host:     "localhost",
	Port:     5432,
	Database: "postgres",

	Username: "postgres",
	Password: "password",
	SSLMode:  "disable",
}
