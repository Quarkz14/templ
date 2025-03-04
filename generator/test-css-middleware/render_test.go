package testcssmiddleware

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

//go:embed expected.html
var expected string

var expectedCSS = `.red_050e5e03{color:red;}
`

func Test(t *testing.T) {
	errs := make([]error, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	// Format the expected value.
	go func() {
		defer wg.Done()
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			errs[0] = fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
	}()

	component := render("Red text")
	h := templ.Handler(component)
	cssmw := templ.NewCSSMiddleware(h, red())

	// Create the actual value.
	var actual string
	go func() {
		defer wg.Done()

		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

		a := new(strings.Builder)
		err := htmlformat.Fragment(a, w.Body)
		if err != nil {
			errs[1] = fmt.Errorf("actual html formatting error: %w", err)
		}
		actual = a.String()
	}()

	var actualCSS string
	go func() {
		defer wg.Done()

		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/styles/templ.css", nil))

		a := new(strings.Builder)
		err := htmlformat.Fragment(a, w.Body)
		if err != nil {
			errs[2] = fmt.Errorf("actual html formatting error: %w", err)
		}
		actualCSS = a.String()
	}()

	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Error(diff)
	}
	if diff := cmp.Diff(expectedCSS, actualCSS); diff != "" {
		t.Error(diff)
	}
}
