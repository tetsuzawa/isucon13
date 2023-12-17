package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func totalTipKey(userID int64) string {
	return fmt.Sprintf("total_tip:%d", userID)
}

func incrTotalTip(ctx context.Context, userID int64, tip int64) error {
	if err := rdb.IncrBy(ctx, totalTipKey(userID), tip).Err(); err != nil {
		return fmt.Errorf("failed to incr total tip: %w", err)
	}
	return nil
}

func decrTotalTip(ctx context.Context, userID int64, tip int64) error {
	if err := rdb.DecrBy(ctx, totalTipKey(userID), tip).Err(); err != nil {
		return fmt.Errorf("failed to decr total tip: %w", err)
	}
	return nil
}

func getTotalTip(ctx context.Context, userID int64) (int64, error) {
	val, err := rdb.Get(ctx, totalTipKey(userID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get total tip: %w", err)
	}
	return val, nil
}

func totalLivecommentsKey(userID int64) string {
	return fmt.Sprintf("total_livecomments:%d", userID)
}

func incrTotalLivecomments(ctx context.Context, userID int64, livecomments int64) error {
	if err := rdb.IncrBy(ctx, totalLivecommentsKey(userID), livecomments).Err(); err != nil {
		return fmt.Errorf("failed to incr total livecomments: %w", err)
	}
	return nil
}

func decrTotalLivecomments(ctx context.Context, userID int64, livecomments int64) error {
	if err := rdb.DecrBy(ctx, totalLivecommentsKey(userID), livecomments).Err(); err != nil {
		return fmt.Errorf("failed to decr total livecomments: %w", err)
	}
	return nil
}

func getTotalLivecomments(ctx context.Context, userID int64) (int64, error) {
	val, err := rdb.Get(ctx, totalLivecommentsKey(userID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get total livecomments: %w", err)
	}
	return val, nil
}
