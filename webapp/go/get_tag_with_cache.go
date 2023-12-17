package main

import (
	"context"
	"fmt"
)

// select(tags): select id from tags where name = ?
// select(tags): select * from tags where id = ?
// select(tags): select * from tags

var tagCache = NewCache[int64, TagModel]()

func SelectAllTagsOnInit(ctx context.Context, db SelectContext) error {
	tags := []TagModel{}
	if err := db.SelectContext(ctx, &tags, "SELECT * FROM tags"); err != nil {
		return err
	}

	for _, tag := range tags {
		tagCache.Set(tag.ID, tag)
	}
	fmt.Println("tagCache:", tagCache.GetAll())
	return nil
}

func SelectAllTags() []*TagModel {
	var tagModels = make([]*TagModel, 0, len(tagCache.GetAll()))
	for _, tag := range tagCache.GetAll() {
		tagModels = append(tagModels, &TagModel{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	return tagModels
}

func GetTagIDsByName(tagName string) []int {
	var tagIDs = make([]int, 0, len(tagCache.GetAll()))
	for tagID, tagModel := range tagCache.GetAll() {
		if tagModel.Name == tagName {
			tagIDs = append(tagIDs, int(tagID))
		}
	}
	fmt.Println("tagIDs:", tagIDs)

	return tagIDs
}

func GetTagByID(tagID int64) TagModel {
	if tagModel, found := tagCache.Get(tagID); found {
		return tagModel
	} else {
		panic("tag not found")
	}
}
