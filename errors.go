package pages

import "errors"

var (
	ErrSiteNotFound     = errors.New("site not found")
	ErrPageNotFound     = errors.New("page not found")
	ErrPageForbidden    = errors.New("page forbidden")
	ErrPageUnauthorized = errors.New("page unauthorized")
	ErrUniqueViolation  = errors.New("unique violation")
)
