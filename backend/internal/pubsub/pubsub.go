package pubsub

import (
	"fmt"
	"sync"
)

type PubSub[T any] struct {
	mu          sync.RWMutex
	subscribers map[string][]chan T // topic -> list of subscriber channels
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		subscribers: make(map[string][]chan T),
	}
}

func (ps *PubSub[T]) Subscribe(topic string) <-chan T {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan T, 10)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	fmt.Printf("Subscribed to topic: %s\n", topic)
	fmt.Printf("Subs count: %d\n", len(ps.subscribers[topic]))
	return ch
}

func (ps *PubSub[T]) Publish(topic string, message T) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.subscribers[topic]; ok {
		for _, ch := range subs {
			select {
			case ch <- message:
				// Message sent successfully
			default:
				fmt.Printf("Warning: Subscriber channel for topic %s is full, dropping message: %v\n", topic, message)
			}
		}
	}
}

func (ps *PubSub[T]) Unsubscribe(topic string, subChan <-chan T) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.subscribers[topic]; ok {
		for i, ch := range subs {
			if ch == subChan {
				close(ch) // Close the channel to signal no more messages
				ps.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				fmt.Printf("Unsubscribed from topic: %s\n", topic)
				fmt.Printf("Subs count: %d\n", len(ps.subscribers[topic]))
				return
			}
		}
	}
}
