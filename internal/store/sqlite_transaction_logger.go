package store

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

type SQLiteTransactionLogger struct {
	db     *sql.DB
	events chan<- Event
	errors <-chan error
}

func (l *SQLiteTransactionLogger) Close() error {
	err := l.db.Close()
	if err != nil {
		return fmt.Errorf("closing sqlite DB: %w", err)
	}

	return nil
}

func (l *SQLiteTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *SQLiteTransactionLogger) WritePut(key string, value string) {
	l.events <- Event{
		Key:   key,
		Value: value,
		Type:  EventPut,
	}
}

func (l *SQLiteTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		Key:  key,
		Type: EventDelete,
	}
}

func NewSQLiteTransactionLogger(filename string, store Store) (*SQLiteTransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	logger := &SQLiteTransactionLogger{
		db: sql.OpenDB(),
	}

	eventsCh, errCh := logger.ReadEvents()
	e, ok := Event{}, true
	for ok && err == nil {
		select {
		case err, ok = <-errCh:
		case e, ok = <-eventsCh:
			switch e.Type {
			case EventDelete:
				err = store.Delete(e.Key)
			case EventPut:
				err = store.Put(e.Key, e.Value)
			}
		}
	}

	logger.Run()

	return logger, nil
}

func (l *SQLiteTransactionLogger) Run() {
	eventCh := make(chan Event, 16)
	l.events = eventCh

	errCh := make(chan error, 1)
	l.errors = errCh

	go func() {
		for e := range eventCh {
			l.currentSequence++
			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.currentSequence, e.Type, e.Key, e.Value)
			if err != nil {
				errCh <- fmt.Errorf("writing event to file: %w", err)
				return
			}
		}
	}()
}

func (l *SQLiteTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	fs := bufio.NewScanner(l.file)
	eventCh := make(chan Event)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		for fs.Scan() {
			var e Event
			line := fs.Text()
			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.Type, &e.Key, &e.Value); err != nil {
				errCh <- fmt.Errorf("scanning line into event: %w", err)
				return
			}

			if l.currentSequence >= e.Sequence {
				errCh <- errors.New("file tx logger out of sequence")
				return
			}

			l.currentSequence = e.Sequence
			eventCh <- e
		}

		if err := fs.Err(); err != nil {
			errCh <- fmt.Errorf("scanning tx log file: %w", err)
			return
		}
	}()

	return eventCh, errCh
}
