package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseContents(t *testing.T) {
	// mock some data
	input := "usr/bin/tool1 pkgA,pkgB\n" +
		"usr/bin/tool2 pkgA\n" +
		"usr/share/doc/tool3 pkgC,pkgA\n" +
		"  \n" +
		"without_spaces_or_packages\n"

	packages, err := parseContents(strings.NewReader(input))
	if err != nil {
		// WTF, this should not happen
		t.Fatalf("unexpected error parsing contents: %v", err)
	}

	expected := map[string]int{
		"pkgA": 3,
		"pkgB": 1,
		"pkgC": 1,
	}

	if len(packages) != len(expected) {
		// again, this should not happen
		t.Fatalf("expected %d packages, got %d", len(expected), len(packages))
	}

	// check one by one,
	// somehthing goes wrong, it will be easier to see what was wrong
	for name, expectedCount := range expected {
		if count, ok := packages[name]; !ok || count != expectedCount {
			t.Errorf("package %q: expected count %d, got %d", name, expectedCount, count)
		}
	}
}

func TestCalculateTopN(t *testing.T) {
	packages := map[string]int{
		"pkgA": 5,
		"pkgB": 20,
		"pkgC": 10,
		"pkgD": 1,
	}

	// validate it orders the list as expected
	t.Run("Top 2 should return highest counts sorted", func(t *testing.T) {
		top := calculateTopN(packages, 2)
		if len(top) != 2 {
			t.Fatalf("expected 2 items, got %d", len(top))
		}
		if top[0].Name != "pkgB" || top[0].Count != 20 {
			t.Errorf("expected first to be pkgB (20), got %s (%d)", top[0].Name, top[0].Count)
		}
		if top[1].Name != "pkgC" || top[1].Count != 10 {
			t.Errorf("expected second to be pkgC (10), got %s (%d)", top[1].Name, top[1].Count)
		}
	})

	// an invalid N, returns empty
	t.Run("Invalid N (<= 0) should return empty slice", func(t *testing.T) {
		top := calculateTopN(packages, 0)
		if len(top) != 0 {
			t.Errorf("expected empty slice for n=0, got %d items", len(top))
		}

		topNeg := calculateTopN(packages, -5)
		if len(topNeg) != 0 {
			t.Errorf("expected empty slice for n=-5, got %d items", len(topNeg))
		}
	})

	// N greater than the length, return the whole list
	// not testing the order, it was cover by the first test
	t.Run("N greater than available items should return all items", func(t *testing.T) {
		top := calculateTopN(packages, 100)
		if len(top) != len(packages) {
			t.Errorf("expected all %d packages, got %d", len(packages), len(top))
		}
	})
}

func TestDownloadContents_Success(t *testing.T) {
	// mock server response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mocked gzip content"))
	}))
	defer server.Close()

	// override the default mirrors, and use the mock server url
	originalMirrors := defaultMirrors
	defaultMirrors = []string{server.URL + "/"}

	defer func() {
		// put the default mirrors back
		defaultMirrors = originalMirrors
	}()

	// give it 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// download the contents
	body, err := downloadContents(ctx, "amd64")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer body.Close()
}

func TestDownloadContents_Failure(t *testing.T) {
	// return not found
	server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server404.Close()

	// override the default mirrors, and use the mock server url
	originalMirrors := defaultMirrors
	defaultMirrors = []string{
		server404.URL + "/",
	}
	defer func() {
		// put the default mirrors back
		defaultMirrors = originalMirrors
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := downloadContents(ctx, "amd64")
	if err == nil {
		t.Fatalf("expected an error when all mirrors fail, got nil")
	}
}

func TestIsValidArchitecture(t *testing.T) {
	tests := []struct {
		arch     string
		expected bool
	}{
		// check some known architectures
		{"amd64", true},
		{"arm64", true},
		{"armel", true},
		{"i386", true},
		// it has to be lowercase, if not it is invalid
		{"AMD64", false},
		// someone trying to do something fishy
		{"../traversal", false},
		{"arch/foo", false},
		// it has to sent something
		{"", false},
		{"    ", false},
		// it has to start with a valid character
		{"-arch", false},
	}

	for _, tc := range tests {
		t.Run(tc.arch, func(t *testing.T) {
			result := isValidArchitecture(tc.arch)
			if result != tc.expected {
				t.Errorf("isValidArchitecture(%q) = %v; expected %v", tc.arch, result, tc.expected)
			}
		})
	}
}
