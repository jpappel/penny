package api

import (
	"net/http"

	"golang.org/x/text/language"
)

func detectLocale(r *http.Request) language.Tag {
	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil || len(tags) == 0 {
		return language.AmericanEnglish
	}

	return tags[0]
}
