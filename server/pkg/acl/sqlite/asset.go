package sqlite

// Grant a user access to read an asset
func (store *Sqlite) GrantUserAssetReadAccess(extUUID string, userUUID string) error {
	return store.GrantUserKindReadAccess("asset", extUUID, userUUID)
}

// Revoke a users access to read an asset
func (store *Sqlite) RevokeUserAssetReadAccess(extUUID string, userUUID string) error {
	return store.RevokeUserKindReadAccess("asset", extUUID, userUUID)
}
func (store *Sqlite) ShareAssetWithPublic(extUUID string) error {
	return store.ShareKindWithPublic("asset", extUUID)
}

// Share an asset only with yourself, this should remove any previous access rules
func (store *Sqlite) MakeAssetPrivate(extUUID string, userUUID string) error {
	return store.MakeKindPrivate("asset", extUUID, userUUID)
}
func (store *Sqlite) DeleteAsset(extUUID string) error {
	return store.DeleteByExtUUID(extUUID)
}

func (store *Sqlite) IsAssetPublic(extUUID string) (bool, error) {
	return store.IsKindPublic("asset", extUUID)
}

func (store *Sqlite) IsAssetPrivate(extUUID string) (bool, error) {
	return store.IsKindPrivate("asset", extUUID)
}

func (store *Sqlite) HasUserAssetReadAccess(extUUID string, userUUID string) (bool, error) {
	return store.HasUserKindReadAccess("asset", extUUID, userUUID)
}
