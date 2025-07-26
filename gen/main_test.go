package main

import "testing"

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// snake_case
		{"first_name", "FirstName"},
		{"user_id", "UserId"},
		{"email_address", "EmailAddress"},
		{"http_server", "HttpServer"},

		// camelCase
		{"firstName", "FirstName"},
		{"userId", "UserId"},
		{"emailAddress", "EmailAddress"},
		{"httpServer", "HttpServer"},

		// PascalCase (already correct)
		{"FirstName", "FirstName"},
		{"UserId", "UserId"},
		{"EmailAddress", "EmailAddress"},
		{"HttpServer", "HttpServer"},

		// Edge cases
		{"", ""},
		{" ", ""},
		{"\n", ""},
		{"_first_name ", "FirstName"},
		{"user__id", "UserId"},
		{" HTTPServer", "HTTPServer"}, // Acronym preservation
		{"xml_http_request", "XmlHttpRequest"},
		{"mixed_SnakeAndCamelCase", "MixedSnakeAndCamelCase"}, // Hybrid case
	}

	for _, test := range tests {
		result := toPascalCase(test.input)
		if result != test.expected {
			t.Errorf("ToPascalCase(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
