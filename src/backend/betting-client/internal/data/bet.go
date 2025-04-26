package data

import "time"

type BetStatus string

const (
	StatusPending BetStatus = "Pending"
	StatusWon     BetStatus = "Won"
	StatusLost    BetStatus = "Lost"
	StatusPaid    BetStatus = "Paid"
	StatusFailed  BetStatus = "Failed"
)

type Bet struct {
	ID                    string    `db:"id"`
	UserID                string    `db:"user_id"`
	EventID               string    `db:"event_id"`
	Amount                float64   `db:"amount"`
	PredictedOutcome      Outcome   `db:"predicted_outcome"`
	RecordedHomeWinChance float64   `db:"recorded_home_win_chance"`
	RecordedAwayWinChance float64   `db:"recorded_away_win_chance"`
	RecordedDrawChance    float64   `db:"recorded_draw_chance"`
	PlacedAt              time.Time `db:"placed_at"`
	Status                BetStatus `db:"status"`
	PayoutAmount          float64   `db:"payout_amount"`
}

type PlaceBetRequest struct {
	UserID           string  `json:"userId" validate:"required,uuid"`
	EventID          string  `json:"eventId" validate:"required,uuid"`
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	PredictedOutcome Outcome `json:"predictedOutcome" validate:"required,oneof=HomeWin AwayWin Draw"`
}

type PayoutNotification struct {
	UserID string  `json:"userId"`
	Amount float64 `json:"amount"`
}

type BetDTO struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	EventID          string    `json:"eventId"`
	Amount           float64   `json:"amount"`
	PredictedOutcome Outcome   `json:"predictedOutcome"`
	PlacedAt         time.Time `json:"placedAt"`
	Status           BetStatus `json:"status"`
}

func MapBetToDTO(b Bet) BetDTO {
	return BetDTO{
		ID:               b.ID,
		UserID:           b.UserID,
		EventID:          b.EventID,
		Amount:           b.Amount,
		PredictedOutcome: b.PredictedOutcome,
		PlacedAt:         b.PlacedAt,
		Status:           b.Status,
	}
}
