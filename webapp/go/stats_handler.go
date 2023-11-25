package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type LivestreamStatistics struct {
	Rank           int64 `json:"rank"`
	ViewersCount   int64 `json:"viewers_count"`
	TotalReactions int64 `json:"total_reactions"`
	TotalReports   int64 `json:"total_reports"`
	MaxTip         int64 `json:"max_tip"`
}

type LivestreamRankingEntry struct {
	LivestreamID int64
	Score        int64
}
type LivestreamRanking []LivestreamRankingEntry

func (r LivestreamRanking) Len() int      { return len(r) }
func (r LivestreamRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r LivestreamRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].LivestreamID < r[j].LivestreamID
	} else {
		return r[i].Score < r[j].Score
	}
}

type UserStatistics struct {
	Rank              int64  `json:"rank"`
	ViewersCount      int64  `json:"viewers_count"`
	TotalReactions    int64  `json:"total_reactions"`
	TotalLivecomments int64  `json:"total_livecomments"`
	TotalTip          int64  `json:"total_tip"`
	FavoriteEmoji     string `json:"favorite_emoji"`
}

type UserRankingEntry struct {
	Username string
	Score    int64
}
type UserRanking []UserRankingEntry

func (r UserRanking) Len() int      { return len(r) }
func (r UserRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r UserRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].Username < r[j].Username
	} else {
		return r[i].Score < r[j].Score
	}
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")
	// ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
	// また、現在の合計視聴者数もだす

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var user UserModel
	if err := tx.GetContext(ctx, &user, "SELECT * FROM users WHERE name = ?", username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusBadRequest, "not found user that has the given username")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
		}
	}

	// ランク算出
	var users []*UserModel
	if err := tx.SelectContext(ctx, &users, "SELECT * FROM users"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get users: "+err.Error())
	}

	userIDStr := strconv.FormatInt(user.ID, 10)
	var rank int64
	if ret := rdb.ZScore(ctx, "ranking", userIDStr); ret.Err() == redis.Nil {
		// nop
	} else if ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get score: "+ret.Err().Error())
	} else {
		score := ret.Val()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert score to int64: "+err.Error())
		}
		scoreStr := fmt.Sprintf("%f.0", score)
		if ret := rdb.ZCount(ctx, "ranking", scoreStr, "+inf"); ret.Err() != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get rank: "+ret.Err().Error())
		}
		rank = int64(ret.Val()) + 1
		if ret := rdb.ZRangeByScore(ctx, "ranking", &redis.ZRangeBy{
			Min: scoreStr, Max: scoreStr,
		}); ret.Err() != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get rank: "+ret.Err().Error())
		} else {
			members := ret.Val()
			sort.Strings(members)
			slices.Reverse(members)
			for _, member := range members {
				if member == userIDStr {
					break
				}
				rank++
			}
		}
	}

	if ret := rdb.ZRevRank(ctx, "ranking", userIDStr); ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get rank: "+ret.Err().Error())
	} else {
		rank = ret.Val() + 1
	}

	// リアクション数
	var totalReactions int64
	if ret := rdb.Get(ctx, "total_reactions:"+userIDStr); ret.Err() == redis.Nil {
		// nop
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get total reactions: "+ret.Err().Error())
	} else {
		totalReactions, err = ret.Int64()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert total reactions to int64: "+err.Error())
		}
	}

	// ライブコメント数、チップ合計
	var totalLivecomments int64
	if ret := rdb.Get(ctx, "total_comments:"+userIDStr); ret.Err() == redis.Nil {
		// nop
	} else if ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get total livecomments: "+ret.Err().Error())
	} else {
		totalLivecomments, err = ret.Int64()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert total livecomments to int64: "+err.Error())
		}
	}
	var totalTip int64
	if ret := rdb.Get(ctx, "total_tip:"+userIDStr); ret.Err() == redis.Nil {
		// nop
	} else if ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get total tip: "+ret.Err().Error())
	} else {
		totalTip, err = ret.Int64()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert total tip to int64: "+err.Error())
		}
	}

	// 合計視聴者数
	var viewersCount int64
	if ret := rdb.Get(ctx, "livestream_viewers:"+userIDStr); ret.Err() == redis.Nil {
		// nop
	} else if ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get viewers count: "+ret.Err().Error())
	} else {
		viewersCount, err = ret.Int64()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert viewers count to int64: "+err.Error())
		}
	}

	// お気に入り絵文字
	var favoriteEmoji string
	if ret := rdb.ZRevRangeWithScores(ctx, "favorite_emoji:"+userIDStr, 0, 10); ret.Err() != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get favorite emojis: "+ret.Err().Error())
	} else {
		rs, err := ret.Result()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get favorite emojis: "+err.Error())
		}
		if len(rs) > 0 {
			var score float64
			favoriteEmoji, score = rs[0].Member.(string), rs[0].Score
			if len(rs) > 1 {
				for _, r := range rs[1:] {
					if r.Score != score {
						break
					}
					member := r.Member.(string)
					if favoriteEmoji < member {
						favoriteEmoji = member
					}
				}
			}
		}
	}

	stats := UserStatistics{
		Rank:              rank,
		ViewersCount:      viewersCount,
		TotalReactions:    totalReactions,
		TotalLivecomments: totalLivecomments,
		TotalTip:          totalTip,
		FavoriteEmoji:     favoriteEmoji,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}
	livestreamID := int64(id)

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var livestream LivestreamModel
	if err := tx.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", livestreamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusBadRequest, "cannot get stats of not found livestream")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
		}
	}

	var livestreams []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}

	// ランク算出
	var ranking LivestreamRanking
	for _, livestream := range livestreams {
		var reactions int64
		if err := tx.GetContext(ctx, &reactions, "SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?", livestream.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count reactions: "+err.Error())
		}

		var totalTips int64
		if err := tx.GetContext(ctx, &totalTips, "SELECT COALESCE(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?", livestream.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tips: "+err.Error())
		}

		score := reactions + totalTips
		ranking = append(ranking, LivestreamRankingEntry{
			LivestreamID: livestream.ID,
			Score:        score,
		})
	}
	sort.Sort(ranking)

	var rank int64 = 1
	for i := len(ranking) - 1; i >= 0; i-- {
		entry := ranking[i]
		if entry.LivestreamID == livestreamID {
			break
		}
		rank++
	}

	// 視聴者数算出
	var viewersCount int64
	if err := tx.GetContext(ctx, &viewersCount, `SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count livestream viewers: "+err.Error())
	}

	// 最大チップ額
	var maxTip int64
	if err := tx.GetContext(ctx, &maxTip, `SELECT COALESCE(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find maximum tip livecomment: "+err.Error())
	}

	// リアクション数
	var totalReactions int64
	if err := tx.GetContext(ctx, &totalReactions, "SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?", livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total reactions: "+err.Error())
	}

	// スパム報告数
	var totalReports int64
	if err := tx.GetContext(ctx, &totalReports, `SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total spam reports: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.JSON(http.StatusOK, LivestreamStatistics{
		Rank:           rank,
		ViewersCount:   viewersCount,
		MaxTip:         maxTip,
		TotalReactions: totalReactions,
		TotalReports:   totalReports,
	})
}
