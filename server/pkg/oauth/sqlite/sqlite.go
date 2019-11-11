package sqlite

import (
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type DatabaseOauth2TokenInfo struct {
	UserUUID     string    `db:"user_uuid"`
	AccessToken  string    `db:"access_token"`
	TokenType    string    `db:"token_type"`
	RefreshToken string    `db:"refresh_token"`
	Expiry       time.Time `db:"expiry"`
}

type Sqlite struct {
	db *sqlx.DB
}

const (
	InsertEntry = `
INSERT INTO oauth2_token_info (user_uuid, access_token, token_type, refresh_token, expiry)
VALUES(?, ?, ?, ?, ?);`

	DeleteByUserUUID = `DELETE FROM oauth2_token_info WHERE user_uuid = ?`
	SelectByUserUUID = `
SELECT
	user_uuid, access_token, token_type, refresh_token, expiry
FROM
	oauth2_token_info
WHERE user_uuid = ?`
)

func NewOAuthReadWriter(db *sqlx.DB) *Sqlite {
	return &Sqlite{
		db: db,
	}
}

func (store *Sqlite) GetTokenInfo(userUUID string) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	var row DatabaseOauth2TokenInfo
	err := store.db.Get(&row, SelectByUserUUID, userUUID)
	if err != nil {
		return token, err
	}

	token.AccessToken = row.AccessToken
	token.TokenType = row.TokenType
	token.RefreshToken = row.RefreshToken
	token.Expiry = row.Expiry
	return token, nil
}

func (store *Sqlite) WriteTokenInfo(userUUID string, token *oauth2.Token) error {
	data := &DatabaseOauth2TokenInfo{
		UserUUID:     userUUID,
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	tx, err := store.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(DeleteByUserUUID, userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(InsertEntry, userUUID, data.AccessToken, data.TokenType, data.RefreshToken, data.Expiry)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
