package events

import (
	"fmt"
	"telegram_bot/client"
	"telegram_bot/db/repository"
	"telegram_bot/db/session"
)

type EventProcess struct {
	tg      *telegram.Client
	storage repository.Storage
	redis   sessions.RedisDB
	offset  int
}

func New(client *telegram.Client, storage repository.Storage, redisdb sessions.RedisDB) *EventProcess {
	return &EventProcess{
		tg:      client,
		storage: storage,
		redis:   redisdb,
	}

}

func (p *EventProcess) Fetch(limit int) ([]Event, error) {

	updates, err := p.tg.GetUpdates(p.offset, limit)

	if err != nil {
		return nil, fmt.Errorf("can't get events: %w", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].UpdateId + 1

	return res, nil
}

func (p *EventProcess) EventHandle(event Event) error {
	return p.doCmd(event)
}

func event(upd telegram.Update) Event {
	res := Event{}
	updType := fetchType(upd)
	switch updType {
	case Message:
		res = Event{
			Text:     fetchMessageText(upd),
			ChatId:   upd.Message.Chat.Id,
			Username: upd.Message.From.Username,
			UpdType:  updType,
		}
	case CallbackQuery:
		res = Event{
			Text:     upd.CallbackQuery.Data,
			ChatId:   upd.CallbackQuery.Message.Chat.Id,
			Username: upd.CallbackQuery.From.Username,
			UpdType:  updType,
		}

	}
	return res
}

func fetchMessageText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}

func fetchCallbackText(upd telegram.Update) string {
	if upd.CallbackQuery == nil {
		return ""
	}
	return upd.CallbackQuery.Message.Text
}

func fetchType(upd telegram.Update) Type {
	if upd.Message == nil {
		if upd.CallbackQuery == nil {
			return Unknown
		} else {
			return CallbackQuery
		}
	}
	return Message
}

//func eventType(evType string) Type2 {
//	switch evType {
//	case "exercisesChoice":
//		return exercisesChoice
//	case "trainChoice":
//		return trainChoice
//	}
//	return UnknownEvent
//}

func convEvent(ev Event) sessions.Session {
	return sessions.Session{
		Username: ev.Username,
		ChatId:   ev.ChatId,
		Message:  []string{ev.Text},
		Train:    sessions.TrainSess{},
	}
}
