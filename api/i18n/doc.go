package i18n

const (
	ValidationWarningLabelToLong         = "The label cannot be longer than 20 characters."
	ValidationWarningLabelNotEmpty       = "The label cannot be empty."
	SuccessAlistNotFound                 = "List not found."
	InternalServerErrorMissingAlistUuid  = "Uuid is missing, possibly an internal error"
	InternalServerErrorMissingUserUuid   = "User.Uuid is missing, possibly an internal error"
	InternalServerErrorTalkingToDatabase = "Issue with talking to the database in %s."
	InputDeleteAlistOperationOwnerOnly   = "Only the owner of the list can remove it."
	DeleteUserLabelSuccess               = "Label %s was removed."
	PostUserLabelJSONFailure             = "Your input is invalid json."
	InputMissingListUuid                 = "The uuid is missing."
	InternalServerErrorDeleteAlist       = "We have failed to remove your list."
)
