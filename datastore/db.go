package datastore

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

const euroId = 1
