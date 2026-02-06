package pages

import (
	"errors"
	"fmt"
)

var (
	ErrSiteNotFound     = errors.New("site not found")
	ErrPageNotFound     = errors.New("page not found")
	ErrPageForbidden    = errors.New("page forbidden")
	ErrPageUnauthorized = errors.New("page unauthorized")
	ErrUniqueViolation  = errors.New("unique violation")
	ErrTemplateEmpty    = errors.New("template is empty")
)

type RedirectError struct {
	url    string
	status int
}

func NewRedirectError(url string, status int) *RedirectError {
	return &RedirectError{url, status}
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("[%d] %s", e.status, e.url)
}

func (e *RedirectError) Redirect() (string, int) {
	return e.url, e.status
}
