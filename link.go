package ghiter

import (
	"strings"
)

type Links []Link

func (l Links) FindByRel(rel string) (Link, bool) {
	for _, link := range l {
		if link.Rel == rel {
			return link, true
		}
	}
	return Link{}, false
}

type Link struct {
	URL    string
	Rel    string
	Params map[string]string
}

func ParseLinkHeader(header string) Links {
	var links Links

	header = strings.TrimSpace(header)
	rawLinks := strings.Split(header, ",")
	for _, l := range rawLinks {
		links = append(links, parseLink(l))
	}

	return links
}

func parseLink(header string) Link {
	header = strings.TrimSpace(header)

	attrs := strings.Split(header, ";")

	rawURL := attrs[0]
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.Trim(rawURL, "<>")

	link := Link{
		URL:    rawURL,
		Params: map[string]string{},
	}

	for i := 1; i < len(attrs); i++ {
		attr := strings.TrimSpace(attrs[i])
		keyVal := strings.Split(attr, "=")
		if len(keyVal) > 1 {
			link.Params[keyVal[0]] = strings.Trim(keyVal[1], `"`)
		}
	}
	link.Rel = link.Params["rel"]

	return link
}
