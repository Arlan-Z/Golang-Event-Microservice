package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventResult defines the possible outcomes of an event.
type EventResult string

const (
	ResultPending  EventResult = "Pending"
	ResultHomeWin  EventResult = "HomeWin"
	ResultAwayWin  EventResult = "AwayWin"
	ResultDraw     EventResult = "Draw"
	ResultCanceled EventResult = "Canceled"
)

// EventType defines the type of sport/game for the event.
type EventType string

const (
	TypeTennis       EventType = "Tennis"
	TypeFootball     EventType = "Football"
	TypeMortalKombat EventType = "MortalKombat"
	TypeCsGo         EventType = "CsGo"
	TypeOther        EventType = "Other"
)

// Определяем ожидаемый формат времени из API
const apiTimeLayout = "2006-01-02T15:04:05" // YYYY-MM-DDTHH:MM:SS

// Event represents the main information about a betting event.
type Event struct {
	ID               string      `json:"id"`
	EventName        string      `json:"eventName"`
	Type             EventType   `json:"type"`
	HomeTeam         string      `json:"homeTeam"`
	AwayTeam         string      `json:"awayTeam"`
	EventStartDate   time.Time   // Убрали тег json
	EventEndDate     time.Time   // Убрали тег json
	EventSubscribers []string    `json:"eventSubscribers"`
	EventResult      EventResult `json:"eventResult,omitempty"`
}

// Кастомный UnmarshalJSON для типа Event
func (e *Event) UnmarshalJSON(data []byte) error {
	type EventAlias struct {
		ID               string      `json:"id"`
		EventName        string      `json:"eventName"`
		Type             EventType   `json:"type"`
		HomeTeam         string      `json:"homeTeam"`
		AwayTeam         string      `json:"awayTeam"`
		EventStartDate   string      `json:"eventStartDate"` // Считываем как строку
		EventEndDate     string      `json:"eventEndDate"`   // Считываем как строку
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

// Round represents details of a single round within an event.
type Round struct {
	RoundNumber   int       `json:"roundNumber"`
	HomeTeamScore int       `json:"homeTeamScore"`
	AwayTeamScore int       `json:"awayTeamScore"`
	RoundDateTime time.Time // Убираем тег json, будем обрабатывать в UnmarshalJSON
}

// --- ДОБАВЛЕНО: Кастомный UnmarshalJSON для типа Round ---
func (r *Round) UnmarshalJSON(data []byte) error {
	// Псевдоним с полем времени как строкой
	type RoundAlias struct {
		RoundNumber   int    `json:"roundNumber"`
		HomeTeamScore int    `json:"homeTeamScore"`
		AwayTeamScore int    `json:"awayTeamScore"`
		RoundDateTime string `json:"roundDateTime"` // Считываем как строку
	}

	var alias RoundAlias
	// Декодируем в псевдоним
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal round alias: %w", err)
	}

	// Копируем обычные поля
	r.RoundNumber = alias.RoundNumber
	r.HomeTeamScore = alias.HomeTeamScore
	r.AwayTeamScore = alias.AwayTeamScore

	// Парсим строку времени, используя наш кастомный формат
	var err error
	r.RoundDateTime, err = time.Parse(apiTimeLayout, alias.RoundDateTime)
	if err != nil {
		return fmt.Errorf("failed to parse RoundDateTime '%s' using layout '%s': %w", alias.RoundDateTime, apiTimeLayout, err)
	}

	return nil
}

// --- КОНЕЦ ДОБАВЛЕННОГО КОДА ---

// EventDetails represents detailed information about an event, including rounds.
// Так как UnmarshalJSON для Round теперь обрабатывает время, EventDetails не требует изменений
type EventDetails struct {
	ID          string  `json:"id"`
	EventName   string  `json:"eventName"`
	EventRounds []Round `json:"eventRounds"`
}

// SubscriptionRequest is the payload for subscribing to an event.
type SubscriptionRequest struct {
	CallbackURL string `json:"callbackUrl"`
}

// SubscriptionResponse is the confirmation message after subscribing.
type SubscriptionResponse string

// EventNotification is the payload received on the callback URL when an event finishes.
type EventNotification struct {
	EventID   string      `json:"eventId"`
	EventName string      `json:"eventName"`
	Result    EventResult `json:"result"`
}
