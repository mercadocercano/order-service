package event

import "time"

// Event representa un evento de dominio
type Event interface {
	GetAggregateID() string
	GetEventType() string
	GetOccurredOn() time.Time
}

// BaseEvent implementación base de un evento
type BaseEvent struct {
	AggregateID string    `json:"aggregate_id"`
	EventType   string    `json:"event_type"`
	OccurredOn  time.Time `json:"occurred_on"`
}

// GetAggregateID retorna el ID del agregado
func (e BaseEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetEventType retorna el tipo de evento
func (e BaseEvent) GetEventType() string {
	return e.EventType
}

// GetOccurredOn retorna cuándo ocurrió el evento
func (e BaseEvent) GetOccurredOn() time.Time {
	return e.OccurredOn
}
