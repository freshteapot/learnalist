package sqlite

func (store *Sqlite) SharePlankHistoryWithPublic(extUUID string) error {
	return store.ShareKindWithPublic("plank/history", extUUID)
}

// Share a plank history only with yourself, this should remove any previous access rules
func (store *Sqlite) MakePlankHistoryPrivate(userUUID string) error {
	return store.MakeKindPrivate("plank/history", userUUID, userUUID)
}

func (store *Sqlite) IsPlankHistoryPublic(extUUID string) (bool, error) {
	return store.IsKindPublic("plank/history", extUUID)
}
