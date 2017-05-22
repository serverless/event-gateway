package endpoints

import "fmt"

// ErrorNotFound occurs when endpoint couldn't been found in the DB.
type ErrorNotFound struct {
	name string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("endpoint %q not found", e.name)
}
