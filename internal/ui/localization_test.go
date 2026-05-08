package ui

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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

func TestTranslatorCachedOrQueueDoesNotBlockOnFetch(t *testing.T) {
	const name = "Slow Player"
	release := make(chan struct{})
	fetchStarted := make(chan struct{})
	translator := &entityTranslatorClient{
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				close(fetchStarted)
				<-release
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`[[["慢速球员","Slow Player",null,null,0]]]`)),
				}, nil
			}),
		},
		cache:    make(map[string]string),
		inflight: make(map[string]*translationCall),
	}

	start := time.Now()
	got := translator.TranslateCachedOrQueue(name)
	if got != name {
		t.Fatalf("expected original name before async translation completes, got %q", got)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("async translation blocked for %s", elapsed)
	}

	select {
	case <-fetchStarted:
	case <-time.After(time.Second):
		t.Fatal("async translation did not start")
	}
	close(release)

	deadline := time.After(time.Second)
	for {
		translator.mu.RLock()
		translated := translator.cache[name]
		translator.mu.RUnlock()
		if translated == "慢速球员" {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("async translation was not cached, got %q", translated)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
