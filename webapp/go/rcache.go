package main

import (
	"context"
	"fmt"
)

func totalTipKey(userID int64) string {
	return fmt.Sprintf("total_tip:%d", userID)
}

func incrTotalTip(ctx context.Context, userID int64) error {
	if err := rdb.Incr(ctx, totalTipKey(userID)).Err(); err != nil {
		return fmt.Errorf("failed to incr total tip: %w", err)
	}
	return nil
}

func decrTotalTip(ctx context.Context, userID int64) error {
	if err := rdb.Decr(ctx, totalTipKey(userID)).Err(); err != nil {
		return fmt.Errorf("failed to decr total tip: %w", err)
	}
	return nil
}

func getTotalTip(ctx context.Context, userID int64) (int64, error) {
	val, err := rdb.Get(ctx, totalTipKey(userID)).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to get total tip: %w", err)
	}
	return val, nil
}
