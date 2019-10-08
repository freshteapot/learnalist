package i18n

const (
	ValidationErrorList                     = "Failed to pass list info. %s"
	ValidationErrorListV1                   = "Failed to pass list type v1."
	ValidationErrorListV2                   = "Failed to pass list type v2."
	ValidationErrorListV3                   = "Failed to pass list type v3."
	ValidationErrorListV4                   = "Failed to pass list type v4."
	ValidationWarningLabelToLong            = "The label cannot be longer than 20 characters."
	ValidationWarningLabelNotEmpty          = "The label cannot be empty."
	ValidationUUIDMismatch                  = "The list uuid in the uri doesnt match that in the payload"
	ValidationAlistTypeV1                   = "Please refer to the documentation on list type v1"
	ValidationAlistTypeV2                   = "Please refer to the documentation on list type v2"
	ValidationAlistTypeV3                   = "Please refer to the documentation on list type v3"
	ValidationAlistTypeV4                   = "Please refer to the documentation on list type v4"
	ValidationUserRegister                  = "Please refer to the documentation on user registration"
	SuccessAlistNotFound                    = "List not found."
	SuccessUserNotFound                     = "User not found."
	InternalServerErrorMissingAlistUuid     = "Uuid is missing, possibly an internal error"
	InternalServerErrorMissingUserUuid      = "User.Uuid is missing, possibly an internal error"
	InternalServerErrorTalkingToDatabase    = "Issue with talking to the database in %s."
	InternalServerErrorAclLookup            = "Issue with talking to the database whilst doing acl lookup"
	InputDeleteAlistOperationOwnerOnly      = "Only the owner of the list can remove it."
	InputSaveAlistOperationOwnerOnly        = "Only the owner of the list can modify it."
	PostUserLabelJSONFailure                = "Your input is invalid json."
	PostShareListJSONFailure                = "Your input is invalid json."
	InputMissingListUuid                    = "The uuid is missing."
	InternalServerErrorDeleteAlist          = "We have failed to remove your list."
	ApiMethodNotSupported                   = "This method is not supported."
	ApiAlistNotFound                        = "Failed to find alist with uuid: %s"
	ApiDeleteAlistSuccess                   = "List %s was removed."
	ApiDeleteUserLabelSuccess               = "Label %s was removed."
	UserInsertAlreadyExistsPasswordNotMatch = "Failed to save."
	UserInsertUsernameExists                = "Username already exists."
	DatabaseLookupNotFound                  = "sql: no rows in result set"
	AclHttpAccessDeny                       = "Access Denied"
)
