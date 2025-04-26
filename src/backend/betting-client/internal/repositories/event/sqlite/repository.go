package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data" // Change path
	"github.com/jmoiron/sqlx"
)

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) FindActiveEvents(ctx context.Context) ([]data.Event, error) {
	events := make([]data.Event, 0)
	query := `SELECT id, event_name, home_team, away_team, home_win_chance, away_win_chance, draw_chance, event_start_date, event_end_date, event_result, type, is_active
              FROM events
              WHERE is_active = 1 AND event_end_date > ?
              ORDER BY event_start_date ASC`

	now := time.Now().UTC()
	err := r.db.SelectContext(ctx, &events, query, now)
	if err != nil {
		if err == sql.ErrNoRows {
			return events, nil
		}
		return nil, fmt.Errorf("error querying active events: %w", err)
	}
	return events, nil
}

func (r *EventRepository) FindByID(ctx context.Context, eventID string) (*data.Event, error) {
	var event data.Event
	query := `SELECT id, event_name, home_team, away_team, home_win_chance, away_win_chance, draw_chance, event_start_date, event_end_date, event_result, type, is_active
              FROM events
              WHERE id = ?`

	err := r.db.GetContext(ctx, &event, query, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying event by ID %s: %w", eventID, err)
	}
	return &event, nil
}

func (r *EventRepository) UpdateResultAndStatus(ctx context.Context, eventID string, result data.Outcome) error {
	query := `UPDATE events SET event_result = ?, is_active = 0 WHERE id = ? AND is_active = 1`
	resultArgs := []interface{}{result, eventID}

	res, err := r.db.ExecContext(ctx, query, resultArgs...)
	if err != nil {
		return fmt.Errorf("error updating event result for %s: %w", eventID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Printf("Warning: failed to get rows affected while updating event %s: %v\n", eventID, err)
	} else if rowsAffected == 0 {
		fmt.Printf("Warning: Update result affected 0 rows for event %s (possibly already inactive or not found)\n", eventID)
	}

	return nil
}
