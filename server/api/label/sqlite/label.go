package sqlite

import (
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/jmoiron/sqlx"
)

type store struct {
	db *sqlx.DB
}

const (
	SqlDeleteLabelByList         = `DELETE FROM alist_labels WHERE alist_uuid=?`
	SqlDeleteLabelByUser         = `DELETE FROM user_labels WHERE user_uuid=? AND label=?`
	SqlDeleteLabelByUserFromList = `DELETE FROM alist_labels WHERE user_uuid=? AND label=?`
	SqlInserListLabel            = "INSERT INTO alist_labels(label, user_uuid, alist_uuid) VALUES (?, ?, ?);"
	SqlInserUserLabel            = `INSERT INTO user_labels(label, user_uuid) VALUES (?, ?);`
	SqlGetListsByUserAndLabel    = `
SELECT
	DISTINCT(alist_uuid)
FROM
	alist_labels
WHERE
	user_uuid=?
AND
	label=?
`
	SqlGetUserLabels = `
SELECT
	label
FROM
	user_labels
WHERE
	user_uuid=?
UNION
SELECT
	label
FROM
	alist_labels
WHERE
	user_uuid=?
`
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

	_, err = store.db.Exec(SqlInserUserLabel, input.Label, input.UserUuid)
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

	_, err = store.db.Exec(SqlInserListLabel, input.Label, input.UserUuid, input.AlistUuid)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return http.StatusOK, nil
		}
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func (store *store) GetUniqueListsByUserAndLabel(label string, user string) ([]string, error) {

	var uuids = []string{}
	err := store.db.Select(&uuids, SqlGetListsByUserAndLabel, user, label)
	if err != nil {
		return uuids, err
	}
	return uuids, nil
}

func (store *store) GetUserLabels(uuid string) ([]string, error) {
	var labels = []string{}

	err := store.db.Select(&labels, SqlGetUserLabels, uuid, uuid)
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
	tx.MustExec(SqlDeleteLabelByUser, user, label)
	tx.MustExec(SqlDeleteLabelByUserFromList, user, label)
	return tx.Commit()
}

func (store *store) RemoveLabelsForAlist(uuid string) error {
	if uuid == "" {
		return nil
	}

	tx := store.db.MustBegin()
	tx.MustExec(SqlDeleteLabelByList, uuid)
	err := tx.Commit()
	return err
}
