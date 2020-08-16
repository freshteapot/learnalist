package sqlite

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/jmoiron/sqlx"
)

type store struct {
	db *sqlx.DB
}

const (
	SQL_LABEL_DELETE_BY_LIST   = `DELETE FROM alist_labels WHERE alist_uuid=?`
	SQL_LABEL_USER_DELETE_USER = `DELETE FROM user_labels WHERE user_uuid=? AND label=?`
	SQL_LABEL_USER_DELETE_LIST = `DELETE FROM alist_labels WHERE user_uuid=? AND label=?`
	SQL_INSERT_LIST_LABEL      = "INSERT INTO alist_labels(label, user_uuid, alist_uuid) VALUES (?, ?, ?);"
)

func NewLabel(db *sqlx.DB) label.LabelReadWriter {
	return &store{
		db: db,
	}
}

func (store *store) PostUserLabel(input label.UserLabel) (int, error) {
	statusCode := http.StatusBadRequest
	err := label.ValidateLabel(input.Label)
	if err != nil {
		return statusCode, err
	}

	query := "INSERT INTO user_labels(label, user_uuid) VALUES (?, ?);"

	_, err = store.db.Exec(query, input.Label, input.UserUuid)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return http.StatusOK, nil
		}
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func (store *store) PostAlistLabel(input label.AlistLabel) (int, error) {
	statusCode := http.StatusBadRequest
	err := label.ValidateLabel(input.Label)
	if err != nil {
		return statusCode, err
	}

	_, err = store.db.Exec(SQL_INSERT_LIST_LABEL, input.Label, input.UserUuid, input.AlistUuid)
	statusCode = http.StatusCreated
	if err != nil {
		statusCode = http.StatusBadRequest
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			statusCode = http.StatusOK
		}
	}
	return statusCode, err
}

func (store *store) GetUniqueListsByUserAndLabel(label string, user string) ([]string, error) {
	queryForUuids := `
SELECT
	DISTINCT(alist_uuid)
FROM
	alist_labels
WHERE
	user_uuid=?
AND
	label=?
`
	fmt.Println("GetUniqueListsByUserAndLabel", label, user)
	var uuids = []string{}
	err := store.db.Select(&uuids, queryForUuids, user, label)
	if err != nil {
		return uuids, err
	}
	return uuids, nil
}

func (store *store) GetUserLabels(uuid string) ([]string, error) {
	var labels = []string{}

	query := `
SELECT label
FROM user_labels
WHERE user_uuid=?
UNION
SELECT label
FROM alist_labels
WHERE user_uuid=?
`
	err := store.db.Select(&labels, query, uuid, uuid)
	if err != nil {
		return labels, err
	}
	return labels, err
}

// Pass in the label and the user (uuid) to remove them from the tables
func (store *store) RemoveUserLabel(label string, user string) error {
	// Update each of them by removin the label in question.
	// TODO MustExec will crash the server
	tx := store.db.MustBegin()
	tx.MustExec(SQL_LABEL_USER_DELETE_USER, user, label)
	tx.MustExec(SQL_LABEL_USER_DELETE_LIST, user, label)
	return tx.Commit()
}

func (store *store) RemoveLabelsForAlist(uuid string) error {
	if uuid == "" {
		return nil
	}

	tx := store.db.MustBegin()
	tx.MustExec(SQL_LABEL_DELETE_BY_LIST, uuid)
	err := tx.Commit()
	return err
}
