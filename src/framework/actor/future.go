package actor

import (
	"errors"
	"time"
)

type Future struct {
	done chan any
	err  chan error
}

func newFuture() *Future {
	return &Future{
		done: make(chan any),
		err:  make(chan error),
	}
}

func (f *Future) Wait() (result any, err error) {
	select {
	case msg := <-f.done:
		return msg, nil
	case err := <-f.err:
		return nil, err
	}
}

func (f *Future) WaitTimeout(timeout time.Duration) (result any, err error) {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case msg := <-f.done:
		return msg, nil
	case err := <-f.err:
		return nil, err
	case <-t.C:
		return nil, errors.New("future wait time out")
	}
}

func (f *Future) respond(response any) {
	select {
	case f.done <- response:
	default:
	}
}

func (f *Future) cancel(err error) {
	select {
	case f.err <- err:
	default:
	}
}
