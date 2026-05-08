package ui

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLocalizeTeamNameFallsBackToFullName(t *testing.T) {
	SetEntityLocalizationEnabled(true)
	t.Cleanup(func() { SetEntityLocalizationEnabled(true) })

	got := localizeTeamName("BHA", "Brighton & Hove Albion")
	if got != "布莱顿" {
		t.Fatalf("expected fallback full-name translation, got %q", got)
	}
}

func TestLocalizeTeamNameKeepsShortNameWhenEntityLocalizationDisabled(t *testing.T) {
	SetEntityLocalizationEnabled(false)
	t.Cleanup(func() { SetEntityLocalizationEnabled(true) })

	got := localizeTeamName("BHA", "Brighton & Hove Albion")
	if got != "BHA" {
		t.Fatalf("expected untranslated short name when entity localization is disabled, got %q", got)
	}
}

func TestTranslatorDoesNotCacheFetchFailures(t *testing.T) {
	const name = "Untranslated Player"
	translator := &entityTranslatorClient{
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		},
		cache:    make(map[string]string),
		inflight: make(map[string]*translationCall),
	}

	got := translator.Translate(name)
	if got != name {
		t.Fatalf("expected original name on fetch failure, got %q", got)
	}
	if _, ok := translator.cache[name]; ok {
		t.Fatalf("fetch failure was cached for %q", name)
	}
}
