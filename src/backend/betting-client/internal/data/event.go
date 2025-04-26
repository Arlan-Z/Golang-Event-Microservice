package data

import "time"

type Outcome string

const (
	HomeWin Outcome = "HomeWin"
	AwayWin Outcome = "AwayWin"
	Draw    Outcome = "Draw"
)

type Event struct {
	ID             string    `db:"id"`
	EventName      string    `db:"event_name"`
	HomeTeam       string    `db:"home_team"`
	AwayTeam       string    `db:"away_team"`
	HomeWinChance  float64   `db:"home_win_chance"`
	AwayWinChance  float64   `db:"away_win_chance"`
	DrawChance     float64   `db:"draw_chance"`
	EventStartDate time.Time `db:"event_start_date"`
	EventEndDate   time.Time `db:"event_end_date"`
	EventResult    *Outcome  `db:"event_result"`
	Type           string    `db:"type"`
	IsActive       bool      `db:"is_active"`
}

type EventDTO struct {
	ID             string    `json:"id"`
	EventName      string    `json:"eventName"`
	HomeTeam       string    `json:"homeTeam"`
	AwayTeam       string    `json:"awayTeam"`
	HomeWinChance  float64   `json:"homeWinChance"`
	AwayWinChance  float64   `json:"awayWinChance"`
	DrawChance     float64   `json:"drawChance"`
	EventStartDate time.Time `json:"eventStartDate"`
	EventEndDate   time.Time `json:"eventEndDate"`
	EventResult    *Outcome  `json:"eventResult,omitempty"`
	Type           string    `json:"type"`
}

func MapEventToDTO(e Event) EventDTO {
	return EventDTO{
		ID:             e.ID,
		EventName:      e.EventName,
		HomeTeam:       e.HomeTeam,
		AwayTeam:       e.AwayTeam,
		HomeWinChance:  e.HomeWinChance,
		AwayWinChance:  e.AwayWinChance,
		DrawChance:     e.DrawChance,
		EventStartDate: e.EventStartDate,
		EventEndDate:   e.EventEndDate,
		EventResult:    e.EventResult,
		Type:           e.Type,
	}
}

func MapEventsToDTOs(events []Event) []EventDTO {
	dtos := make([]EventDTO, len(events))
	for i, e := range events {
		dtos[i] = MapEventToDTO(e)
	}
	return dtos
}
