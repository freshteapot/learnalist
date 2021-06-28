package sqlite

// Grant a user access to read an plankHistory
func (store *Sqlite) GrantUserPlankHistoryReadAccess(extUUID string, userUUID string) error {
	return store.GrantUserKindReadAccess("plank/history", extUUID, userUUID)
}

// Revoke a users access to read an plankHistory
func (store *Sqlite) RevokeUserPlankHistoryReadAccess(extUUID string, userUUID string) error {
	return store.RevokeUserKindReadAccess("plank/history", extUUID, userUUID)
}
func (store *Sqlite) SharePlankHistoryWithPublic(extUUID string) error {
	return store.ShareKindWithPublic("plank/history", extUUID)
}

// Share a plank history only with yourself, this should remove any previous access rules
func (store *Sqlite) MakePlankHistoryPrivate(userUUID string) error {
	return store.MakeKindPrivate("plank/history", userUUID, userUUID)
}

func (store *Sqlite) HasUserPlankHistoryReadAccess(extUUID string, userUUID string) (bool, error) {
	return store.HasUserKindReadAccess("plank/history", extUUID, userUUID)
}

func (store *Sqlite) IsPlankHistoryPublic(extUUID string) (bool, error) {
	return store.IsKindPublic("plank/history", extUUID)
}
