package static

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goroute/route"
)

type (
	// Options defines the config for Static middleware.
	Options struct {
		// Skipper defines a function to skip middleware.
		Skipper route.Skipper

		// Root directory from where the static content is served.
		// Required.
		Root string `yaml:"root"`

		// Index file for serving a directory.
		// Optional. Default value "index.html".
		Index string `yaml:"index"`

		// Enable HTML5 mode by forwarding all not-found requests to root so that
		// SPA (single-page application) can handle the routing.
		// Optional. Default value false.
		HTML5 bool `yaml:"html5"`

		// Enable directory browsing.
		// Optional. Default value false.
		Browse bool `yaml:"browse"`
	}
)

type Option func(*Options)

func GetDefaultOptions() Options {
	return Options{
		Skipper: route.DefaultSkipper,
		Root:    ".",
		Index:   "index.html",
		HTML5:   false,
		Browse:  false,
	}
}

func Skipper(skipper route.Skipper) Option {
	return func(o *Options) {
		o.Skipper = skipper
	}
}

func Root(root string) Option {
	return func(o *Options) {
		o.Root = root
	}
}

func Index(index string) Option {
	return func(o *Options) {
		o.Index = index
	}
}

func HTML5(html5 bool) Option {
	return func(o *Options) {
		o.HTML5 = html5
	}
}

func Browse(browse bool) Option {
	return func(o *Options) {
		o.Browse = browse
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>{{ .Name }}</title>
  <style>
    body {
			font-family: Menlo, Consolas, monospace;
			padding: 48px;
		}
		header {
			padding: 4px 16px;
			font-size: 24px;
		}
    ul {
			list-style-type: none;
			margin: 0;
    	padding: 20px 0 0 0;
			display: flex;
			flex-wrap: wrap;
    }
    li {
			width: 300px;
			padding: 16px;
		}
		li a {
			display: block;
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
			text-decoration: none;
			transition: opacity 0.25s;
		}
		li span {
			color: #707070;
			font-size: 12px;
		}
		li a:hover {
			opacity: 0.50;
		}
		.dir {
			color: #E91E63;
		}
		.file {
			color: #673AB7;
		}
  </style>
</head>
<body>
	<header>
		{{ .Name }}
	</header>
	<ul>
		{{ range .Files }}
		<li>
		{{ if .Dir }}
			{{ $name := print .Name "/" }}
			<a class="dir" href="{{ $name }}">{{ $name }}</a>
			{{ else }}
			<a class="file" href="{{ .Name }}">{{ .Name }}</a>
			<span>{{ .Size }}</span>
		{{ end }}
		</li>
		{{ end }}
  </ul>
</body>
</html>
`

// New returns a Static middleware.
func New(options ...Option) route.MiddlewareFunc {
	// Apply options.
	opts := GetDefaultOptions()
	for _, opt := range options {
		opt(&opts)
	}

	// Index template
	t, err := template.New("index").Parse(html)
	if err != nil {
		panic(fmt.Sprintf("static: %v", err))
	}

	return func(c route.Context, next route.HandlerFunc) (err error) {
		if opts.Skipper(c) {
			return next(c)
		}

		p := c.Request().URL.Path
		if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
			p = c.Param("*")
		}
		p, err = url.PathUnescape(p)
		if err != nil {
			return
		}
		name := filepath.Join(opts.Root, path.Clean("/"+p)) // "/"+ for security

		fi, err := os.Stat(name)
		if err != nil {
			if os.IsNotExist(err) {
				if err = next(c); err != nil {
					if he, ok := err.(*route.HTTPError); ok {
						if opts.HTML5 && he.Code == http.StatusNotFound {
							return c.File(filepath.Join(opts.Root, opts.Index))
						}
					}
					return
				}
			}
			return
		}

		if fi.IsDir() {
			index := filepath.Join(name, opts.Index)
			fi, err = os.Stat(index)

			if err != nil {
				if opts.Browse {
					return listDir(t, name, c.Response())
				}
				if os.IsNotExist(err) {
					return next(c)
				}
				return
			}

			return c.File(index)
		}

		return c.File(name)
	}
}

func listDir(t *template.Template, name string, res *route.Response) (err error) {
	file, err := os.Open(name)
	if err != nil {
		return
	}
	files, err := file.Readdir(-1)
	if err != nil {
		return
	}

	// Create directory index.
	res.Header().Set(route.HeaderContentType, route.MIMETextHTMLCharsetUTF8)
	data := struct {
		Name  string
		Files []interface{}
	}{
		Name: name,
	}
	for _, f := range files {
		data.Files = append(data.Files, struct {
			Name string
			Dir  bool
			Size string
		}{f.Name(), f.IsDir(), formatFileSize(f.Size())})
	}
	return t.Execute(res, data)
}

const (
	_ = 1.0 << (10 * iota) // Ignore first value by assigning to blank identifier.
	KB
	MB
	GB
	TB
	PB
	EB
)

func formatFileSize(b int64) string {
	multiple := ""
	value := float64(b)

	switch {
	case b >= EB:
		value /= EB
		multiple = "EB"
	case b >= PB:
		value /= PB
		multiple = "PB"
	case b >= TB:
		value /= TB
		multiple = "TB"
	case b >= GB:
		value /= GB
		multiple = "GB"
	case b >= MB:
		value /= MB
		multiple = "MB"
	case b >= KB:
		value /= KB
		multiple = "KB"
	case b == 0:
		return "0"
	default:
		return strconv.FormatInt(b, 10) + "B"
	}

	return fmt.Sprintf("%.2f%s", value, multiple)
}
