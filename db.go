package mint

import "database/sql"

//Database connection intercase
type Database interface {
	Connection() *sql.DB
}
