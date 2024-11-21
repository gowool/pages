package pages

import (
	"database/sql"
	"errors"
)

var (
	ErrInternal     = errors.New("internal server error")
	ErrSiteNotFound = errors.New("site not found")
	ErrPageNotFound = errors.New("page not found")
	ErrMenuNotFound = errors.New("menu not found")
)

func IsOneOfNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrSiteNotFound) || errors.Is(err, ErrPageNotFound) || errors.Is(err, ErrMenuNotFound)
}
