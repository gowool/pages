package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"net"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gowool/pages"
	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/repository"
)

type HybridPageConfig struct {
	Skipper       middleware.Skipper
	PageHandler   pages.PageHandler
	CfgRepository repository.Configuration
}

func HybridPage(cfg HybridPageConfig) echo.MiddlewareFunc {
	if cfg.PageHandler == nil {
		panic("page handler is not specified")
	}
	if cfg.CfgRepository == nil {
		panic("configuration repository is not specified")
	}
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	bPool := sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}

	rPool := sync.Pool{
		New: func() any { return new(Response) },
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()

			if cfg.Skipper(c) ||
				pages.SkipSelectSite(r.Context()) ||
				pages.SkipSelectPage(r.Context()) ||
				pages.IsAjax(r) {
				return next(c)
			}

			configuration, err := cfg.CfgRepository.Load(r.Context())
			if err != nil {
				return err
			}

			if configuration.IgnorePattern(r.Pattern) || configuration.IgnoreURI(r.URL.Path) {
				return next(c)
			}

			page := pages.CtxPage(r.Context())
			if page == nil {
				return repository.ErrPageNotFound
			}

			if !page.IsHybrid() || !page.Decorate {
				return next(c)
			}

			buffer := bPool.Get().(*bytes.Buffer)
			buffer.Reset()
			defer bPool.Put(buffer)

			response := rPool.Get().(*Response)
			defer func() {
				response.Reset(nil, nil)
				rPool.Put(response)
			}()

			w := c.Response()
			response.Reset(w.Writer, buffer)
			w.Writer = response

			if err = next(c); err != nil {
				w.Writer = response.Writer
				return err
			}
			w.Writer = response.Writer
			w.Committed = false

			if !pages.IsTextHTML(c.Response().Header()) || response.Status != http.StatusOK || pages.PageNotDecorate(c.Response()) {
				if response.Committed || response.Buffer.Len() > 0 {
					_, err = w.Write(response.Buffer.Bytes())
				}
				return err
			}

			data := pages.CtxData(r.Context())
			data["content"] = template.HTML(internal.String(buffer.Bytes()))
			ctx := pages.WithData(r.Context(), data)
			c.SetRequest(r.WithContext(ctx))

			return cfg.PageHandler.Handle(c)
		}
	}
}

var ErrHeaderAlreadyCommitted = errors.New("response already committed")

type Response struct {
	Writer    http.ResponseWriter
	Buffer    *bytes.Buffer
	Status    int
	Size      int64
	Committed bool
}

func (r *Response) StatusCode() int {
	return r.Status
}

func (r *Response) WriteHeader(status int) {
	if r.Committed {
		panic(ErrHeaderAlreadyCommitted)
	}
	r.Status = status
	r.Committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	return r.Buffer.Write(b)
}

// Header returns the header map for the writer that will be sent by
// WriteHeader. Changing the header after a call to WriteHeader (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to WriteHeader (see example)
// To suppress implicit response headers, set their value to nil.
// Example: https://golang.org/pkg/net/http/#example_ResponseWriter_trailers
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

// Pusher implements the http.Pusher interface to allow an HTTP handler to
// constructs a synthetic request using the given target and options.
// See [http.Pusher](https://golang.org/pkg/net/http/#Pusher)
func (r *Response) Pusher() http.Pusher {
	if pusher, ok := r.Writer.(http.Pusher); ok {
		return pusher
	}
	return nil
}

// Unwrap returns the original http.ResponseWriter.
// ResponseController can be used to access the original http.ResponseWriter.
// See [https://go.dev/blog/go1.20]
func (r *Response) Unwrap() http.ResponseWriter {
	return r.Writer
}

func (r *Response) Reset(w http.ResponseWriter, b *bytes.Buffer) {
	r.Writer = w
	r.Buffer = b
	r.Size = 0
	r.Status = http.StatusOK
	r.Committed = false
}
