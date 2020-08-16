package label

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
)

type LabelReadWriter interface {
	// Labels
	PostUserLabel(label UserLabel) (int, error)
	RemoveUserLabel(label string, uuid string) error
	PostAlistLabel(label AlistLabel) (int, error)
	GetUserLabels(uuid string) ([]string, error)
	GetUniqueListsByUserAndLabel(label string, user string) ([]string, error)
	RemoveLabelsForAlist(uuid string) error
}

type UserLabel struct {
	Label    string
	UserUuid string
}

type AlistLabel struct {
	Label     string
	UserUuid  string
	AlistUuid string
}

func NewUserLabel(label string, user string) UserLabel {
	userLabel := UserLabel{
		Label:    label,
		UserUuid: user,
	}
	return userLabel
}

func NewAlistLabel(label string, user string, alist string) AlistLabel {
	alistLabel := AlistLabel{
		Label:     label,
		UserUuid:  user,
		AlistUuid: alist,
	}
	return alistLabel
}

func ValidateLabel(label string) error {
	if label == "" {
		return errors.New(i18n.ValidationWarningLabelNotEmpty)
	}

	if len(label) > 20 {
		return errors.New(i18n.ValidationWarningLabelToLong)
	}
	return nil
}
