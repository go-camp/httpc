package pathx

import (
	"fmt"
	"net/url"
	"testing"
)

func TestTemplate_Execute(t *testing.T) {
	testCases := []struct {
		Name string

		Path string
		Data map[string]string

		ExpectPath    string
		ExpectRawPath string
	}{
		{
			Name: "empty",
		},

		{
			Name: "slash",

			Path: "/",

			ExpectPath:    "/",
			ExpectRawPath: "/",
		},

		{
			Name: "slot",

			Path: "/:store/products/:product.json",
			Data: map[string]string{
				"store":   "1001",
				"product": "2020",
			},

			ExpectPath:    "/1001/products/2020.json",
			ExpectRawPath: "/1001/products/2020.json",
		},

		{
			Name: "escape",

			Path: "/:name",
			Data: map[string]string{
				"name": "中/",
			},

			ExpectPath:    "/中/",
			ExpectRawPath: "/%E4%B8%AD%2F",
		},

		{
			Name: "ext",

			Path: "/:product.:ext",
			Data: map[string]string{
				"product": "2020",
				"ext":     "xml",
			},

			ExpectPath:    "/2020.xml",
			ExpectRawPath: "/2020.xml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			pathTemplate, err := ParseTemplate(tc.Path)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			path, rawPath := pathTemplate.Execute(tc.Data)

			unescapePath, err := url.PathUnescape(rawPath)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			if path != unescapePath {
				t.Fatalf("expect unescaped path is %q, got %q", path, unescapePath)
			}

			if tc.ExpectPath != path {
				t.Fatalf("expect path is %q, got %q", tc.ExpectPath, path)
			}
			if tc.ExpectRawPath != rawPath {
				t.Fatalf("expect raw path is %q, got %q", tc.ExpectRawPath, rawPath)
			}
		})
	}
}

func ExampleTemplate_Resolve() {
	baseURL, err := url.Parse("https://www.example.com?foo=bar#/path")
	if err != nil {
		fmt.Printf("Parse Base URL failed: %v\n", err)
		return
	}

	pathTemplate := MustParseTemplate("/:store/products/:product.json")
	resolvedURL := pathTemplate.Resolve(baseURL, map[string]string{
		"store":   "10001",
		"product": "2020",
	})
	fmt.Println("Resolved URL:", resolvedURL)
	// Output:
	// Resolved URL: https://www.example.com/10001/products/2020.json?foo=bar#/path
}
