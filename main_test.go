package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestAnalyzePage(t *testing.T) {
	// Sample minimal HTML
	sampleHTML := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Page</title>
	</head>
	<body>
		<h1>Welcome</h1>
		<h2>Subheading</h2>
		<a href="/internal">Internal Link</a>
		<a href="https://external.com">External Link</a>
		<form action="/login">
			<input type="password" name="pass" />
		</form>
	</body>
	</html>`

	doc, err := html.Parse(strings.NewReader(sampleHTML))
	if err != nil {
		t.Fatalf("Failed to parse sample HTML: %v", err)
	}

	result := analyzePage(doc, "http://example.com")

	// Assertions
	if result.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", result.Title)
	}

	if result.HeadingsCount["h1"] != 1 {
		t.Errorf("Expected 1 h1 heading, got %d", result.HeadingsCount["h1"])
	}

	if result.HeadingsCount["h2"] != 1 {
		t.Errorf("Expected 1 h2 heading, got %d", result.HeadingsCount["h2"])
	}

	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}

	if !result.HasLoginForm {
		t.Errorf("Expected login form to be detected")
	}
}
