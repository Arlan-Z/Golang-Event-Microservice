package data

import (
	"fmt"
	"time"
)

type ExternalEventDTO struct {
	APIEventID      string   `json:"id"`
	Name            string   `json:"eventName"`
	TeamHome        string   `json:"homeTeam"`
	TeamAway        string   `json:"awayTeam"`
	CoefficientHome *float64 `json:"homeWinChance"`
	CoefficientAway *float64 `json:"awayWinChance"`
	CoefficientDraw *float64 `json:"drawChance"`
	StartsAt        string   `json:"eventStartDate"`
	EndsAt          string   `json:"eventEndDate"`
	SportType       string   `json:"type"`
	Result          *string  `json:"eventResult"`
}

func MapExternalToInternalEvent(ext ExternalEventDTO) (Event, error) {
	internalEvent := Event{
		ID:            ext.APIEventID,
		EventName:     ext.Name,
		HomeTeam:      ext.TeamHome,
		AwayTeam:      ext.TeamAway,
		Type:          ext.SportType,
		HomeWinChance: 0,
		AwayWinChance: 0,
		DrawChance:    0,
		IsActive:      true,
	}

	if ext.CoefficientHome != nil {
		internalEvent.HomeWinChance = *ext.CoefficientHome
	}
	if ext.CoefficientAway != nil {
		internalEvent.AwayWinChance = *ext.CoefficientAway
	}
	if ext.CoefficientDraw != nil {
		internalEvent.DrawChance = *ext.CoefficientDraw
	}

	const timeLayout = "2006-01-02T15:04:05"

	startTime, err := time.Parse(timeLayout, ext.StartsAt)
	if err != nil {
		return Event{}, fmt.Errorf("failed to parse eventStartDate '%s': %w", ext.StartsAt, err)
	}
	internalEvent.EventStartDate = startTime.UTC()

	endTime, err := time.Parse(timeLayout, ext.EndsAt)
	if err != nil {
		return Event{}, fmt.Errorf("failed to parse eventEndDate '%s': %w", ext.EndsAt, err)
	}
	internalEvent.EventEndDate = endTime.UTC()

	if ext.Result != nil && *ext.Result != "" {
		var outcome Outcome
		switch *ext.Result {
		case "HomeWin":
			outcome = HomeWin
		case "AwayWin":
			outcome = AwayWin
		case "Draw":
			outcome = Draw
		default:
			return Event{}, fmt.Errorf("unknown eventResult format '%s'", *ext.Result)
		}
		internalEvent.EventResult = &outcome
		internalEvent.IsActive = false
	}

	if internalEvent.EventResult == nil && time.Now().UTC().After(internalEvent.EventEndDate) {
		internalEvent.IsActive = false
	}

	return internalEvent, nil
}
