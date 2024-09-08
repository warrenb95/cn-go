package store

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

type EventType uint32

const (
	_ EventType = iota
	EventPut
	EventDelete
)

type FileTransactionLogger struct {
	events          chan<- Event
	errors          <-chan error
	currentSequence uint64
	file            *os.File
}

type Event struct {
	Sequence uint64
	Type     EventType
	Key      string
	Value    string
}

func (l *FileTransactionLogger) WritePut(key string, value string) {
	l.events <- Event{
		Key:   key,
		Value: value,
		Type:  EventPut,
	}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		Key:  key,
		Type: EventPut,
	}
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &FileTransactionLogger{
		file: file,
	}, nil
}

func (l *FileTransactionLogger) Run() {
	eventCh := make(chan Event, 16)
	l.events = eventCh

	errCh := make(chan error, 1)
	l.errors = errCh

	go func() {
		for e := range eventCh {
			l.currentSequence++
			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s", l.currentSequence, e.Type, e.Key, e.Value)
			if err != nil {
				errCh <- fmt.Errorf("writing event to file: %w", err)
				return
			}
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	fs := bufio.NewScanner(l.file)
	eventCh := make(chan Event)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		var e Event
		for fs.Scan() {
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
