package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventResult string

const (
	ResultPending  EventResult = "Pending"
	ResultHomeWin  EventResult = "HomeWin"
	ResultAwayWin  EventResult = "AwayWin"
	ResultDraw     EventResult = "Draw"
	ResultCanceled EventResult = "Canceled"
)

type EventType string

const (
	TypeTennis       EventType = "Tennis"
	TypeFootball     EventType = "Football"
	TypeMortalKombat EventType = "MortalKombat"
	TypeCsGo         EventType = "CsGo"
	TypeOther        EventType = "Other"
)

const apiTimeLayout = "2006-01-02T15:04:05"

type Event struct {
	ID               string    `json:"id"`
	EventName        string    `json:"eventName"`
	Type             EventType `json:"type"`
	HomeTeam         string    `json:"homeTeam"`
	AwayTeam         string    `json:"awayTeam"`
	EventStartDate   time.Time
	EventEndDate     time.Time
	EventSubscribers []string    `json:"eventSubscribers"`
	EventResult      EventResult `json:"eventResult,omitempty"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type EventAlias struct {
		ID               string      `json:"id"`
		EventName        string      `json:"eventName"`
		Type             EventType   `json:"type"`
		HomeTeam         string      `json:"homeTeam"`
		AwayTeam         string      `json:"awayTeam"`
		EventStartDate   string      `json:"eventStartDate"`
		EventEndDate     string      `json:"eventEndDate"`
		EventSubscribers []string    `json:"eventSubscribers"`
		EventResult      EventResult `json:"eventResult,omitempty"`
	}

	var alias EventAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal event alias: %w", err)
	}

	e.ID = alias.ID
	e.EventName = alias.EventName
	e.Type = alias.Type
	e.HomeTeam = alias.HomeTeam
	e.AwayTeam = alias.AwayTeam
	e.EventSubscribers = alias.EventSubscribers
	e.EventResult = alias.EventResult

	var err error
	e.EventStartDate, err = time.Parse(apiTimeLayout, alias.EventStartDate)
	if err != nil {
		return fmt.Errorf("failed to parse EventStartDate '%s' using layout '%s': %w", alias.EventStartDate, apiTimeLayout, err)
	}

	e.EventEndDate, err = time.Parse(apiTimeLayout, alias.EventEndDate)
	if err != nil {
		return fmt.Errorf("failed to parse EventEndDate '%s' using layout '%s': %w", alias.EventEndDate, apiTimeLayout, err)
	}

	return nil
}

type Round struct {
	RoundNumber   int       `json:"roundNumber"`
	HomeTeamScore int       `json:"homeTeamScore"`
	AwayTeamScore int       `json:"awayTeamScore"`
	RoundDateTime time.Time `json:"roundDateTime"`
}

type EventDetails struct {
	ID          string  `json:"id"`
	EventName   string  `json:"eventName"`
	EventRounds []Round `json:"eventRounds"`
}

type SubscriptionRequest struct {
	CallbackURL string `json:"callbackUrl"`
}

type SubscriptionResponse string

type EventNotification struct {
	EventID   string      `json:"eventId"`
	EventName string      `json:"eventName"`
	Result    EventResult `json:"result"`
}
