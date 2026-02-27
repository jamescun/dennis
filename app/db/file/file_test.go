package file

import (
	"github.com/jamescun/dennis/app/db"
)

// ensure DB implements the db.DB interface.
var _ db.DB = (*DB)(nil)
