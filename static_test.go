package static

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goroute/route"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	mux := route.NewServeMux()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := mux.NewContext(req, rec)

	h := New(Root("testdata"))(route.NotFoundHandler)

	assert := assert.New(t)
	if assert.NoError(h(c)) {
		assert.Contains(rec.Body.String(), "Route")
	}
}

func TestStaticFileFound(t *testing.T) {
	mux := route.NewServeMux()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := mux.NewContext(req, rec)

	h := New(Root("testdata"))(route.NotFoundHandler)

	assert := assert.New(t)
	req = httptest.NewRequest(http.MethodGet, "/images/walle.png", nil)
	rec = httptest.NewRecorder()
	c = mux.NewContext(req, rec)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(rec.Header().Get(route.HeaderContentLength), "219885")
	}
}

func TestStaticFileNotFound(t *testing.T) {
	mux := route.NewServeMux()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := mux.NewContext(req, rec)

	h := New(Root("testdata"))(route.NotFoundHandler)

	assert := assert.New(t)
	req = httptest.NewRequest(http.MethodGet, "/none", nil)
	rec = httptest.NewRecorder()
	c = mux.NewContext(req, rec)
	he := h(c).(*route.HTTPError)
	assert.Equal(http.StatusNotFound, he.Code)
}

func TestStaticHTML5(t *testing.T) {
	mux := route.NewServeMux()
	req := httptest.NewRequest(http.MethodGet, "/random", nil)
	rec := httptest.NewRecorder()
	c := mux.NewContext(req, rec)

	h := New(Root("testdata"), HTML5(true))(route.NotFoundHandler)

	assert := assert.New(t)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "Route")
	}
}

func TestStaticBrowse(t *testing.T) {
	mux := route.NewServeMux()
	req := httptest.NewRequest(http.MethodGet, "/file1.txt", nil)
	rec := httptest.NewRecorder()
	c := mux.NewContext(req, rec)

	h := New(
		Root("testdata/browse"),
		Browse(true),
	)(route.NotFoundHandler)

	assert := assert.New(t)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "Hello")
	}
}
