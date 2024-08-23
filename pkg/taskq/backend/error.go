package backend

import "errors"

var (
	ErrNotSupported     = errors.New("not supported")
	ErrTaskNameRequired = errors.New("taskq: Job.TaskName is required")
)
