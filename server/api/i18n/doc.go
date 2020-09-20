package i18n

import (
	"errors"
)

const (
	ValidationErrorList                         = "Failed to pass list info. %s" // TODO maybe remove
	ValidationErrorListV1                       = "Failed to pass list type v1."
	ValidationErrorListV2                       = "Failed to pass list type v2."
	ValidationErrorListV3                       = "Failed to pass list type v3."
	ValidationErrorListV4                       = "Failed to pass list type v4."
	ValidationWarningLabelToLong                = "The label cannot be longer than 20 characters."
	ValidationUUIDMismatch                      = "The list uuid in the uri doesnt match that in the payload"
	ValidationAlists                            = "Please refer to the documentation on lists"
	ValidationAlistTypeV1                       = "Please refer to the documentation on list type v1"
	ValidationAlistTypeV2                       = "Please refer to the documentation on list type v2"
	ValidationAlistTypeV3                       = "Please refer to the documentation on list type v3"
	ValidationAlistTypeV4                       = "Please refer to the documentation on list type v4"
	ValidationUserRegister                      = "Please refer to the documentation on user registration"
	ValidationLabel                             = "Please refer to the documentation on label(s)"
	SuccessAlistNotFound                        = "List not found."
	SuccessUserNotFound                         = "User not found."
	InternalServerErrorMissingAlistUuid         = "Uuid is missing, possibly an internal error"
	InternalServerErrorMissingUserUuid          = "User.Uuid is missing, possibly an internal error"
	InternalServerErrorTalkingToDatabase        = "Issue with talking to the database in %s."
	InternalServerErrorAclLookup                = "Issue with talking to the database whilst doing acl lookup"
	InternalServerErrorFunny                    = "Sadly, our service has taken a nap."
	InputDeleteAlistOperationOwnerOnly          = "Only the owner of the list can remove it."
	InputSaveAlistOperationOwnerOnly            = "Only the owner of the list can modify it."
	InputSaveAlistOperationFromRestriction      = "List not shareable due to restriction on the info.from.kind"
	InputSaveAlistOperationFromModify           = "Unable to modify info.from"
	InputSaveAlistOperationFromKindNotSupported = "The kind used in info, is not supported"
	PostUserLabelJSONFailure                    = "Your input is invalid json."
	InputAlistJSONFailure                       = "Your input is invalid json."
	PostShareListJSONFailure                    = "Your input is invalid json."
	InputLogoutJSONFailure                      = "Your input is invalid json."
	InputMissingListUuid                        = "The uuid is missing."
	InternalServerErrorDeleteAlist              = "We have failed to remove your list."
	ApiMethodNotSupported                       = "This method is not supported."
	ApiAlistNotFound                            = "Failed to find alist with uuid: %s"
	ApiDeleteAlistSuccess                       = "List %s was removed."
	ApiDeleteUserLabelSuccess                   = "Label %s was removed."
	UserInsertAlreadyExistsPasswordNotMatch     = "Failed to save."
	UserInsertUsernameExists                    = "Username already exists."
	AclHttpAccessDeny                           = "Access Denied"
	ApiShareYouCantShareWithYourself            = "Today, we dont let you share with yourself"
	ApiShareValidationError                     = "Please refer to the documentation on sharing a list"
	ApiShareListSuccessWithPublic               = "List is now public"
	ApiShareListSuccessWithFriends              = "List is now private to the owner and those granted access"
	ApiShareListSuccessPrivate                  = "List is now private to the owner"
	ApiShareReadAccessInvalidWithNotShared      = "You cant grant or revoke read access when the list is shared as private"
	ApiShareNoChange                            = "No change made"
	ApiUserLogoutError                          = "Please refer to the api documentation regarding /user/logout"
	ApiUserLoginError                           = "Please refer to the api documentation regarding /user/login"
)

var (
	ErrorCannotReadResponse                     = errors.New("Cannot read response.")
	ErrorInternal                               = errors.New("An internal error has occurred. If you see this repeatedly, please contact support.")
	ErrorUserSessionActivate                    = errors.New("challenge doesnt exist or is active")
	ErrorUserAlreadyExists                      = errors.New("user.already.exists")
	ErrorUserAlreadyExistsWrongPassword         = errors.New("user.already.exists.wrong.password")
	ErrorInputSaveAlistFromKindNotSupported     = errors.New(InputSaveAlistOperationFromKindNotSupported)
	ErrorInputSaveAlistOperationFromRestriction = errors.New(InputSaveAlistOperationFromRestriction)
	ErrorInputSaveAlistOperationFromModify      = errors.New(InputSaveAlistOperationFromModify)
	ErrorInputSaveAlistOperationOwnerOnly       = errors.New(InputSaveAlistOperationOwnerOnly)
	ErrorAListFromDomainMisMatch                = errors.New("In info.from, the domain doesnt match the kind")
	ErrorValidationWarningLabelNotEmpty         = errors.New("The label cannot be empty.")
	ErrorListNotFound                           = errors.New("list.not.found")
)
