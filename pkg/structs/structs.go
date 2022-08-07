package structs

import (
	"time"
)

type EmptyCtx struct {
}

func (e *EmptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (e *EmptyCtx) Done() <-chan struct{} {
	return nil
}

func (e *EmptyCtx) Err() error {
	return nil
}

func (e *EmptyCtx) Value(key any) any {
	return nil
}

func (e *EmptyCtx) String() string {
	return ""
}