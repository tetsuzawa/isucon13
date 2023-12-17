package main

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func initScore(ctx context.Context, db sqlx.ExecerContext, userModel UserModel) error {
	if _, err := db.ExecContext(ctx, "INSERT INTO user_scores (user_id, name, reactions, tip) VALUES (?, ?, 0, 0)", userModel.ID, userModel.Name); err != nil {
		return fmt.Errorf("failed to init score: %w", err)
	}
	return nil
}

func incrReactionsToScore(ctx context.Context, db sqlx.ExecerContext, userID int64, reactions int64) error {
	if _, err := db.ExecContext(ctx, "UPDATE user_scores SET reactions = reactions + ? WHERE user_id = ?", reactions, userID); err != nil {
		return fmt.Errorf("failed to incr reactions to score: %w", err)
	}
	return nil
}

func incrTipToScore(ctx context.Context, db sqlx.ExecerContext, userID int64, tip int64) error {
	if _, err := db.ExecContext(ctx, "UPDATE user_scores SET tip = tip + ? WHERE user_id = ?", tip, userID); err != nil {
		return fmt.Errorf("failed to incr tip to score: %w", err)
	}
	return nil
}

func decrTipToScore(ctx context.Context, db sqlx.ExecerContext, userID int64, tip int64) error {
	if _, err := db.ExecContext(ctx, "UPDATE user_scores SET tip = tip - ? WHERE user_id = ?", tip, userID); err != nil {
		return fmt.Errorf("failed to decr tip to score: %w", err)
	}
	return nil
}
