package models

import (
	"github.com/freshteapot/learnalist-api/api/uuid"
)

func NewLabel() *Label {
	label := &Label{
		Uuid: uuid.GetUUID("label"),
	}
	return label
}

func (dal *DAL) GetLabel(Uuid string) (*Label, error) {
	label := Label{}
	query := `
SELECT *
FROM label
WHERE uuid=$1
`
	err := dal.Db.Get(&label, query, Uuid)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (dal *DAL) GetLabelsByUser(Uuid string) []Label {
	labels := []Label{}
	query := `
SELECT l.*, al.alist_uuid
FROM label as l
LEFT JOIN alist_labels as al ON l.uuid=al.label_uuid
WHERE user_uuid=$1
`
	rows, _ := dal.Db.Queryx(query, Uuid)
	for rows.Next() {
		label := Label{}
		rows.StructScan(&label)
		labels = append(labels, label)
	}
	return labels
}

// Save the label to the database
// If the AlistUuid is set, it will also save this.
func (dal *DAL) SaveLabel(label Label) error {
	var err error
	labelInsertQuery := "INSERT INTO label (uuid, label, user_uuid) VALUES (:uuid, :label, :user_uuid)"
	labelLinkInsert := "INSERT INTO alist_labels (alist_uuid, label_uuid) VALUES (:alist_uuid, :label_uuid)"

	link := &AlistLabelLink{
		AlistUuid: label.AlistUuid,
		LabelUuid: label.Uuid,
	}

	_, err = dal.Db.NamedExec(labelInsertQuery, label)
	if err != nil {
		if err.Error() != "UNIQUE constraint failed: label.label, label.user_uuid" {
			return err
		}
	}

	if label.AlistUuid != "" {
		_, err = dal.Db.NamedExec(labelLinkInsert, link)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dal *DAL) RemoveLabel(uuid string) error {
	query1 := `
DELETE
FROM label
WHERE uuid=$1
`
	query2 := `
DELETE
FROM alist_labels
WHERE label_uuid=$1
`
	tx := dal.Db.MustBegin()
	tx.MustExec(query1, uuid)
	tx.MustExec(query2, uuid)
	err := tx.Commit()
	return err
}
