package router

import (
	"time"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
)

// WaitForFunction returns a chan that is closed when a function is created.
// Primarily for testing purposes.
func (router *Router) WaitForFunction(space string, id function.ID) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.Function(space, id)
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForAsyncSubscriber returns a chan that is closed when an event has a subscriber.
// Primarily for testing purposes.
func (router *Router) WaitForAsyncSubscriber(method, path string, eventType eventpkg.TypeName) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.AsyncSubscribers(method, path, eventType)
			if len(res) > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForSyncSubscriber returns a chan that is closed when an a sync subscriber is created.
// Primarily for testing purposes.
func (router *Router) WaitForSyncSubscriber(method, path string, eventType eventpkg.TypeName) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			subscriber := router.targetCache.SyncSubscriber(method, path, eventType)
			if subscriber != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForEventType returns a chan that is closed when a event type is created.
// Primarily for testing purposes.
func (router *Router) WaitForEventType(space string, name eventpkg.TypeName) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.EventType(space, name)
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}
