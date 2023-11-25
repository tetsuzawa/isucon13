package main

import (
	"context"
)

var themeCache = NewCache[int64, ThemeModel]()

func GetThemeWithCache(ctx context.Context, db GetContext, userID int64) (ThemeModel, error) {
	if themeModel, found := themeCache.Get(userID); found {
		return themeModel, nil
	}
	themeModel := ThemeModel{}
	if err := db.GetContext(ctx, &themeModel, "SELECT * FROM themes WHERE id = ?", userID); err != nil {
		return ThemeModel{}, err
	}
	themeCache.Set(userID, themeModel)
	return themeModel, nil
}
