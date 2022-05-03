package controllers

import (
	"context"
	"fmt"
)

type Credentials struct {
	Login    string
	Password string
}

type CtxKey string

var UserCTXKey CtxKey = "ctxUserID"

func UserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserCTXKey).(int)
	if !ok {
		return -1, fmt.Errorf("no user in context")
	}

	return userID, nil
}
