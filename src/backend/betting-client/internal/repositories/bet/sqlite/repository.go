package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Arlan-Z/def-betting-api/internal/data" // Change path
	"github.com/jmoiron/sqlx"
)

type BetRepository struct {
	db *sqlx.DB
}

func NewBetRepository(db *sqlx.DB) *BetRepository {
	return &BetRepository{db: db}
}

func (r *BetRepository) Save(ctx context.Context, bet *data.Bet) error {
	query := `INSERT INTO bets (id, user_id, event_id, amount, predicted_outcome,
                        recorded_home_win_chance, recorded_away_win_chance, recorded_draw_chance,
                        placed_at, status, payout_amount)
              VALUES (:id, :user_id, :event_id, :amount, :predicted_outcome,
                      :recorded_home_win_chance, :recorded_away_win_chance, :recorded_draw_chance,
                      :placed_at, :status, :payout_amount)`

	_, err := r.db.NamedExecContext(ctx, query, bet)
	if err != nil {
		return fmt.Errorf("error saving bet: %w", err)
	}
	return nil
}

func (r *BetRepository) FindPendingByEventID(ctx context.Context, eventID string) ([]data.Bet, error) {
	bets := make([]data.Bet, 0)
	query := `SELECT id, user_id, event_id, amount, predicted_outcome, recorded_home_win_chance, recorded_away_win_chance, recorded_draw_chance, placed_at, status, payout_amount
              FROM bets
              WHERE event_id = ? AND status = ?`

	err := r.db.SelectContext(ctx, &bets, query, eventID, data.StatusPending)
	if err != nil {
		if err == sql.ErrNoRows {
			return bets, nil
		}
		return nil, fmt.Errorf("error querying pending bets for event %s: %w", eventID, err)
	}
	return bets, nil
}

func (r *BetRepository) UpdateStatusAndPayout(ctx context.Context, betID string, status data.BetStatus, payout float64) error {
	query := `UPDATE bets SET status = ?, payout_amount = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, payout, betID)
	if err != nil {
		return fmt.Errorf("error updating bet status for %s: %w", betID, err)
	}
	return nil
}

func (r *BetRepository) UpdateStatus(ctx context.Context, betID string, status data.BetStatus) error {
	query := `UPDATE bets SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, betID)
	if err != nil {
		return fmt.Errorf("error updating bet status for %s: %w", betID, err)
	}
	return nil
}
