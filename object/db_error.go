package object

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	mssql "github.com/microsoft/go-mssqldb"
	"modernc.org/sqlite"
)

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	var postgresErr *pq.Error
	if errors.As(err, &postgresErr) && postgresErr.Code == "23505" {
		return true
	}

	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		code := sqliteErr.Code()
		if code == 1555 || code == 2067 {
			return true
		}
	}

	var mssqlErr mssql.Error
	if errors.As(err, &mssqlErr) {
		if mssqlErr.Number == 2601 || mssqlErr.Number == 2627 {
			return true
		}
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint failed") ||
		strings.Contains(message, "violates unique constraint") ||
		strings.Contains(message, "unique key constraint")
}
