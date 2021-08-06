package main

import (
	"fmt"
	"os"
)

const (
	errAuth = "wrong user name or password"
)

var (
	pgHost, pgDB, pgUser, pgPassword string
	svcPort, svcUser, svcPassword    string
)

func init() {

	// load DBMS params
	pgHost = os.Getenv("PG_HOST")
	pgDB = os.Getenv("PG_DB")
	pgUser = os.Getenv("PG_USER")
	pgPassword = os.Getenv("PG_PASSWORD")

	// load service params
	svcPort = os.Getenv("SVC_PORT")
	svcUser = os.Getenv("SVC_USER")
	svcPassword = os.Getenv("SVC_PASSWORD")

}

func dbConnectSting() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s)", pgUser, pgPassword, pgHost, pgDB)
}
