package event_consumer

import (
	"log"
	"telegram_bot/events"
	"time"
)

type ConsumerEvent struct {
	fetcher   events.Fetcher
	processor events.EventHandler
	batchSize int
}

func New(fetcher events.Fetcher, processor events.EventHandler, batchSize int) ConsumerEvent {
	return ConsumerEvent{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c *ConsumerEvent) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("err consumer %s", err)
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)
			continue
		}

	}
}

func (c *ConsumerEvent) handleEvents(events []events.Event) error {
	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		if err := c.processor.EventHandle(event); err != nil {
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
}
