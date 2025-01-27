package api

import (
	"net/http"
	"time"

	"golang.org/x/text/language"
)

func detectLocale(r *http.Request) language.Tag {
	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil || len(tags) == 0 {
		return language.AmericanEnglish
	}

	return tags[0]
}

func localizeTimestamp(epochTime int64, timezone string) (time.Time, error) {
	return time.Time{}, nil
}
