package dbtestutils

import "fmt"

type ConnectionParams struct {
	Host string
	Port int

	Database string

	Username string
	Password string

	SSLMode string
}

// Turns connection params into postgresql connection string.
func (cp *ConnectionParams) ConnString() string {
	connstr := fmt.Sprintf(
		"user='%s' password='%s' host='%s' port='%d' dbname='%s' sslmode='%s'",
		cp.Username, cp.Password, cp.Host, cp.Port, cp.Database, cp.SSLMode,
	)

	return connstr
}

// TODO(ville): Read these variables from env or something.
func ConnectionParamsDefaults() *ConnectionParams {
	return &ConnectionParams{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",

		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
	}
}
