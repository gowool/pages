package pages

import "errors"

var (
	ErrSiteNotFound    = errors.New("site not found")
	ErrPageNotFound    = errors.New("page not found")
	ErrPrivatePage     = errors.New("page is private")
	ErrUniqueViolation = errors.New("unique violation")
)
