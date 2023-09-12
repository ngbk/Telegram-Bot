package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type EventHandler interface {
	EventHandle(e Event) error
}

type Event struct {
	Text     string
	ChatId   int
	Username string
	UpdType  Type
}

type Type int

type Type2 int

const (
	exercisesChoice Type2 = iota
	trainChoice
	UnknownEvent
)

const (
	Unknown Type = iota
	Message
	CallbackQuery
)
