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
		return Event{}, fmt.Errorf("не удалось распарсить eventStartDate '%s': %w", ext.StartsAt, err)
	}
	internalEvent.EventStartDate = startTime.UTC()

	endTime, err := time.Parse(timeLayout, ext.EndsAt)
	if err != nil {
		return Event{}, fmt.Errorf("не удалось распарсить eventEndDate '%s': %w", ext.EndsAt, err)
	}
	internalEvent.EventEndDate = endTime.UTC()

	makeInactive := false

	if ext.Result != nil && *ext.Result != "" {
		var outcome *Outcome

		switch *ext.Result {
		case "HomeWin":
			res := HomeWin
			outcome = &res
			makeInactive = true
		case "AwayWin":
			res := AwayWin
			outcome = &res
			makeInactive = true
		case "Draw":
			res := Draw
			outcome = &res
			makeInactive = true
		case "Canceled":
			makeInactive = true
		case "Pending":
			break
		default:
			return Event{}, fmt.Errorf("unknown eventResult format '%s'", *ext.Result)
		}

		if outcome != nil {
			internalEvent.EventResult = outcome
		}
	}

	if makeInactive {
		internalEvent.IsActive = false
	} else {
		if time.Now().UTC().After(internalEvent.EventEndDate) {
			internalEvent.IsActive = false
		}
	}

	return internalEvent, nil
}
