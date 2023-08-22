package db

import (
	"context"
	"fmt"
	"pg-to-es/internal/config"
	"sync"

	"github.com/lib/pq"
)

type Listener struct {
	cfg         config.Pg
	lstnr       *pq.Listener
	deltaStream chan string
	closeOnce   sync.Once
}

// Initialize Listener
func NewListener(cfg config.Pg) (*Listener, error) {
	var listenerErr error
	listener := pq.NewListener(cfg.String(), cfg.ListenerMinReconnectInterval,
		cfg.ListenerMaxReconnectInterval, func(event pq.ListenerEventType, err error) {
			if err != nil {
				listenerErr = fmt.Errorf("pq.NewListener() failed, err: %s", err)
			}
		})
	if listenerErr != nil {
		return nil, listenerErr
	}
	return &Listener{
		cfg:         cfg,
		lstnr:       listener,
		deltaStream: make(chan string),
	}, listenerErr
}

// Start Listening to CRUD operations
func (l *Listener) Start(ctx context.Context) (<-chan string, error) {
	err := l.lstnr.Listen(l.cfg.ListenerChannel)
	if err != nil {
		return nil, err
	}
	go l.listen(ctx)
	return l.deltaStream, nil
}

// Stop listening
func (l *Listener) Stop() {
	l.closeOnce.Do(func() {
		close(l.deltaStream)
		l.lstnr.Close()
	})
}

func (l *Listener) listen(ctx context.Context) {
	for {
		select {
		case n := <-l.lstnr.Notify:
			l.deltaStream <- n.Extra
		case <-ctx.Done():
			return
		}
	}
}
