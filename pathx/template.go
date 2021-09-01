package pathx

import (
	"net/url"
	"regexp"
	"strings"
)

type Template struct {
	text  string
	slots [][2]int
}

var slotRe = regexp.MustCompile(`:[a-zA-Z0-9_-]+`)

// ParseTemplate parses text as a template.
//
// Valid text examples:
//   /:store/products/:product.:ext
//   /:store.json
//   /:name
func ParseTemplate(text string) (*Template, error) {
	t := &Template{text: text}

	matches := slotRe.FindAllStringSubmatchIndex(text, -1)
	t.slots = make([][2]int, len(matches))
	for i, match := range matches {
		t.slots[i] = [2]int{match[0], match[1]}
	}

	return t, nil
}

// MustParseTemplate returns a parsed template or panic.
func MustParseTemplate(text string) *Template {
	t, err := ParseTemplate(text)
	if err != nil {
		panic(err)
	}
	return t
}

// Execute applies a parsed template to the specified data object.
//
// path is same as Path field of url.URL.
// rawPath is same RawPath field of url.URL.
func (t *Template) Execute(data map[string]string) (path, rawPath string) {
	var pathB, rawPathB strings.Builder
	pathB.Grow(len(t.text))
	rawPathB.Grow(len(t.text))

	n := 0
	for _, slot := range t.slots {
		s, e := slot[0], slot[1]

		lit := t.text[n:s]
		pathB.WriteString(lit)
		rawPathB.WriteString(lit)

		val := data[t.text[s+1:e]]
		pathB.WriteString(val)
		rawPathB.WriteString(url.PathEscape(val))

		n = e
	}
	tail := t.text[n:]
	pathB.WriteString(tail)
	rawPathB.WriteString(tail)

	return pathB.String(), rawPathB.String()
}

// Resolve applies the parsed template to the specified data object
// and resolves returned path to an absolute URL from an absolute baseURL.
func (t *Template) Resolve(baseURL *url.URL, data map[string]string) *url.URL {
	path, rawPath := t.Execute(data)
	pathURL := &url.URL{Path: path, RawPath: rawPath}
	if baseURL == nil {
		return pathURL
	}

	resolvedURL := baseURL.ResolveReference(pathURL)
	resolvedURL.RawQuery = baseURL.RawQuery
	resolvedURL.Fragment = baseURL.Fragment
	resolvedURL.RawFragment = baseURL.RawFragment

	return resolvedURL
}

func (t *Template) String() string {
	return t.text
}
