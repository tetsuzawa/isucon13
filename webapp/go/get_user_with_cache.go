package main

import (
	"context"
)

type GetContext interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

var userCache = NewCache[int64, UserModel]()

func GetUserWithCache(ctx context.Context, db GetContext, userID int64) (*UserModel, error) {
	if userModel, found := userCache.Get(userID); found {
		return &userModel, nil
	}
	userModel := UserModel{}
	if err := db.GetContext(ctx, &userModel, "SELECT * FROM users WHERE id = ?", userID); err != nil {
		return nil, err
	}
	userCache.Set(userID, userModel)
	return &userModel, nil
}
