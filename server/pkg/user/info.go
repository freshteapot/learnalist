package user

type userInfo struct {
	storage UserInfoRepository
}

func NewUserInfo(storage UserInfoRepository) UserInfoRepository {
	return userInfo{
		storage: storage,
	}
}

func (r userInfo) Get(userUUID string) (UserPreference, error) {
	return r.storage.Get(userUUID)
}

func (r userInfo) Save(userUUID string, pref UserPreference) error {
	return r.storage.Save(userUUID, pref)
}
