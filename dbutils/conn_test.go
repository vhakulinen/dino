package dbutils_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vhakulinen/dino/dbutils"
)

// TODO(ville): Read these variables from env or something.
var defaultConnectionParams = dbutils.ConnectionParams{
	Host:     "localhost",
	Port:     5432,
	Database: "postgres",

	Username: "postgres",
	Password: "password",
	SSLMode:  "disable",
}

func TestConnectionParams_ConnString(t *testing.T) {
	params := defaultConnectionParams

	got := params.ConnString()
	expected := "user='postgres' password='password' host='localhost' port='5432' dbname='postgres' sslmode='disable'"

	if diff := cmp.Diff(got, expected); diff != "" {
		t.Fatal(diff)
	}
}
