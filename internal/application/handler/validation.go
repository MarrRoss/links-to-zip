package handler

import "net/url"

func IsValidURL(link string) (bool, url.URL) {
	parsedLink, err := url.Parse(link)
	if err != nil || parsedLink.Path == "" {
		return false, url.URL{}
	}
	return true, *parsedLink
}
