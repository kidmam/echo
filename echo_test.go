package echo_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	. "github.com/webx-top/echo"
	test "github.com/webx-top/echo/testing"
)

func request(method, path string, e *Echo) (int, string) {
	rec := test.Request(method, path, e)
	return rec.Code, rec.Body.String()
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-1")
			return next.Handle(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next.Handle(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next.Handle(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next.Handle(c)
		}
	})

	// Route
	e.Get("/", func(c Context) error {
		return c.String("OK")
	})

	c, b := request(GET, "/", e)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return errors.New("error")
		}
	})
	e.Get("/", NotFoundHandler)
	c, _ := request(GET, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestGroupMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-1")
			return next.Handle(c)
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("0")
			return next.Handle(c)
		}
	})

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-3")
			return next.Handle(c)
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-2")
			return next.Handle(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next.Handle(c)
		}
	})

	g := e.Group("/", func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next.Handle(c)
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next.Handle(c)
		}
	})

	// Route
	g.Get("", func(c Context) error {
		return c.String("OK")
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("4")
			return next.Handle(c)
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("5")
			return next.Handle(c)
		}
	})

	c, b := request(GET, "/", e)
	assert.Equal(t, "-3-2-1012345", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)

	buf = new(bytes.Buffer)
	e.RebuildRouter()
	c, b = request(GET, "/", e)
	assert.Equal(t, "-3-2-1012345", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.Get("/ok", func(c Context) error {
		return c.String("OK")
	})
	e.Get("/view/:id", func(c Context) error {
		return c.String(c.Param(`id`))
	})
	e.Get("/file/*", func(c Context) error {
		return c.String(c.P(0))
	})

	c, b := request(GET, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
	c, b = request(GET, "/view/123", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "123", b)
	c, b = request(POST, "/view/0", e)
	assert.Equal(t, http.StatusMethodNotAllowed, c)
	assert.Equal(t, "Method Not Allowed", b)
	c, b = request(GET, "/file/path/to/file.js", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "path/to/file.js", b)
}

func TestEchoMeta(t *testing.T) {
	e := New()

	g := e.Group("/root")

	g.Get("/", e.MetaHandler(
		H{"version": 1.0, "data": H{"by": "handler"}},
		func(c Context) error {
			return c.JSON(c.Route().Meta)
		},
	))

	var meta H

	for _, route := range e.Routes() {
		if route.Path == "/root/" {
			meta = route.Meta
		}
	}
	expected := H{
		"version": 1.0,
		"data": H{
			"by": "handler",
		},
	}
	assert.Equal(t, expected, meta)

	c, b := request(GET, "/root/", e)
	assert.Equal(t, http.StatusOK, c)
	expected2, _ := json.Marshal(expected)
	assert.Equal(t, string(expected2), b)
}

func TestEchoData(t *testing.T) {
	data := NewData(nil)
	data.SetCode(0)
	assert.Equal(t, 0, data.Code.Int())
	assert.Equal(t, `0`, fmt.Sprintf(`%d`, data.Code))
	assert.Equal(t, `Failure`, fmt.Sprintf(`%v`, data.Code))
	assert.Equal(t, `Failure`, fmt.Sprintf(`%s`, data.Code))
	assert.Equal(t, `Failure`, data.State)
}
