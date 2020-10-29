package sqlite

// Grant a user access to read an challenge
func (store *Sqlite) GrantUserChallengeWriteAccess(extUUID string, userUUID string) error {
	return store.GrantUserKindWriteAccess("challenge", extUUID, userUUID)
}

// Revoke a users access to read an challenge
func (store *Sqlite) RevokeUserChallengeWriteAccess(extUUID string, userUUID string) error {
	return store.RevokeUserKindWriteAccess("challenge", extUUID, userUUID)
}
func (store *Sqlite) ShareChallengeWithPublic(extUUID string) error {
	return store.ShareKindWithPublic("challenge", extUUID)
}

// Share an challenge only with yourself, this should remove any previous access rules
func (store *Sqlite) MakeChallengePrivate(extUUID string, userUUID string) error {
	return store.MakeKindPrivate("challenge", extUUID, userUUID)
}
func (store *Sqlite) DeleteChallenge(extUUID string) error {
	return store.DeleteByExtUUID(extUUID)
}

func (store *Sqlite) HasUserChallengeWriteAccess(extUUID string, userUUID string) (bool, error) {
	return store.HasUserKindWriteAccess("challenge", extUUID, userUUID)
}
