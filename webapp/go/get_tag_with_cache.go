package main

// select(tags): select id from tags where name = ?
// select(tags): select * from tags where id = ?
// select(tags): select * from tags

var tagCache = NewCache[int64, TagModel]()
