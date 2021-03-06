package markdown

import (
	"bytes"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Markdown{})
}

// Markdown is a middleware which render response bodies as markdown.
type Markdown struct {
	// The template by which to render markdown. Default is "simple".
	Template string `json:"template,omitempty"`

	// The MIME types for which to render markdown. It is important to use
	// this if the route matchers do not exclude images or other binary files.
	// Default is text/markdown.
	MIMETypes []string `json:"mime_types,omitempty"`
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// CaddyModule returns the Caddy module information.
func (Markdown) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.markdown",
		New: func() caddy.Module { return new(Markdown) },
	}
}

// Provision provisions md.
func (md *Markdown) Provision(ctx caddy.Context) error {
	caddy.Log().Info("Provision:", zap.String("template", md.Template))
	if md.MIMETypes == nil {
		md.MIMETypes = []string{"text/markdown"}
	}
	if md.Template == "" {
		md.Template = "simple"
	}
	return nil
}

// Validate ensures md has a valid configuration.
func (md *Markdown) Validate() error {
	return nil
}

func (md *Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	caddy.Log().Info("ServeHTTP:", zap.String("template", md.Template), zap.String("path", r.URL.Path))
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	// shouldBuf determines whether to render markdown on this response,
	// since generally we will not want to render for images or CSS, etc.
	shouldBuf := func(status int, header http.Header) bool {
		ct := header.Get("Content-Type")
		for _, mt := range md.MIMETypes {
			if strings.Contains(ct, mt) {
				return true
			}
		}
		return false
	}

	rec := caddyhttp.NewResponseRecorder(w, buf, shouldBuf)

	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	if !rec.Buffered() {
		return nil
	}

	body, err := renderMarkdown(buf.String())
	if err != nil {
		return caddyhttp.Error(http.StatusInternalServerError, err)
	}

	tmpl, ok := templates[md.Template]
	if !ok {
		// if not a built-in template, try as resource file
		buf.Reset()
		fs := http.Dir(".")
		file, err := fs.Open(md.Template)
		if err == nil {
			defer file.Close()
			io.Copy(buf, file)
		}
		if buf.Len() > 0 {
			tmpl = buf.String()
		} else {
			tmpl = "{{.Body}}"
		}
	}

	orignalRequest := r.Context().Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
	html := strings.Replace(tmpl, "{{.Title}}", path.Base(orignalRequest.URL.Path), 1)
	html = strings.Replace(html, "{{.Body}}", body, 1)

	buf.Reset()
	buf.WriteString(html)

	rec.Header().Set("Content-Type", "text/html; charset=utf-8")
	rec.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	rec.Header().Del("Accept-Ranges") // we don't know ranges for dynamically-created content
	rec.Header().Del("Last-Modified") // useless for dynamic content since it's always changing

	// we don't know a way to quickly generate etag for dynamic content,
	// and weak etags still cause browsers to rely on it even after a
	// refresh, so disable them until we find a better way to do this
	rec.Header().Del("Etag")

	return rec.WriteResponse()
}

func renderMarkdown(inputStr string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(), // TODO: this is not awesome, maybe should be configurable?
		),
	)

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	err := md.Convert([]byte(inputStr), buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Markdown)(nil)
	_ caddy.Validator             = (*Markdown)(nil)
	_ caddyhttp.MiddlewareHandler = (*Markdown)(nil)
)
