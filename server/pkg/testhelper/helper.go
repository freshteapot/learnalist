package testhelper

import (
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func GetMockDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(fmt.Sprintf("an error '%s' was not expected when opening a stub database connection", err.Error()))
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, err
}
