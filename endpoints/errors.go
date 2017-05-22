package endpoints

import "fmt"

// ErrorNotFound occurs when endpoint couldn't been found in the DB.
type ErrorNotFound struct {
	name string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("endpoint %q not found", e.name)
}

// ErrorTargetNotFound occurs when requested target couldn't been found.
type ErrorTargetNotFound struct {
	name string
}

func (e ErrorTargetNotFound) Error() string {
	return fmt.Sprintf("endpoint %q doesn't have specified target", e.name)
}
