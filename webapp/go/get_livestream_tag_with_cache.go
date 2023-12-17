package main

import (
	"context"
)

type SelectContext interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

var livestreamTagCache = NewCache[int64, []*LivestreamTagModel]()

func GetLivestreamTagWithCache(ctx context.Context, db SelectContext, livestreamID int64) ([]*LivestreamTagModel, error) {
	if livesStreamTagModels, found := livestreamTagCache.Get(livestreamID); found {
		return livesStreamTagModels, nil
	}
	livestreamTagModels := []*LivestreamTagModel{}
	if err := db.SelectContext(ctx, &livestreamTagModels, "SELECT * FROM livestream_tags WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	livestreamTagCache.Set(livestreamID, livestreamTagModels)
	return livestreamTagModels, nil
}
