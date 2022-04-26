package repo

import "errors"

var (
	ErrUserNotFound      = errors.New("user does not exist")
	ErrUserAuthFailed    = errors.New("user authentication failed")
	ErrUserAlreadyExists = errors.New("duplicate user name")

	ErrDuplicateOrder                    = errors.New("duplicate order")
	ErrOrderAlreadyUploadedByCurrentUser = errors.New("order already exist for this user")
	ErrOrderCreatedByAnotherUser         = errors.New("order already exist for another user")

	ErrInternalError = errors.New("internal error")
)
