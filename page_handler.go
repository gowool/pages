package pages

import (
	"fmt"

	"github.com/gowool/wo"
)

const (
	PageCMSPattern    = "/{_page_cms...}"
	HomeHybridPattern = "/{$}"
)

func PageHandler[T Resolver]() func(T) error {
	return func(e T) error {
		if !e.HasSite() {
			return fmt.Errorf("page handler: %w", ErrSiteNotFound)
		}

		if !e.HasPage() {
			return fmt.Errorf("page handler: %w", ErrPageNotFound)
		}

		page := e.Page()

		for key, values := range page.Header {
			for i, value := range values {
				if i == 0 {
					e.Response().Header().Set(key, value)
				} else {
					e.Response().Header().Add(key, value)
				}
			}
		}

		if page.Template == "" {
			return e.NoContent(e.Status())
		}

		ct := e.Response().Header().Get(wo.HeaderContentType)
		if ct == "" {
			ct = wo.MIMETextHTMLCharsetUTF8
		}

		return e.Render(e.Status(), ct, page.Template)
	}
}
