package web

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// NegotiateContentType parses the "Accept" header off a request and returns the mutually
// negotiated content type based on a list of available content types.
//
// If the "Accept" header is missing or empty, the first available content type will
// be selected as the negotiated content type.
//
// An empty string will be returned if no match was found, and should then
// result in a 406 (Not Acceptable) response to the client.
//
// An error will be returned if the request accept header is malformed.
func NegotiateContentType(req *http.Request, availableContentTypes ...string) (string, error) {
	if len(availableContentTypes) == 0 {
		return "", errNegotiateContentTypeAvailableEmpty
	}
	mediaTypes, err := parseAccept(req.Header.Get(HeaderAccept))
	if err != nil {
		return "", err
	}
	if len(mediaTypes) == 0 {
		return availableContentTypes[0], nil
	}

	bestQ := -1.0
	bestWild := 3
	var bestContentType, availableType, availableSubtype string
	for _, available := range availableContentTypes {
		availableType, availableSubtype = splitOn(available, rune('/'))
		if availableType == "" || availableSubtype == "" {
			return "", errNegotiateContentTypeInvalidAvailableContentType
		}
		for _, mediaType := range mediaTypes {
			switch {
			case mediaType.Q == 0:
				continue
			case (mediaType.Q < bestQ):
				continue
			case mediaType.Type == "*" && mediaType.Subtype == "*":
				if mediaType.Q > bestQ || bestWild > 2 {
					bestQ = mediaType.Q
					bestWild = 2
					bestContentType = available
				}
				continue
			case strings.EqualFold(mediaType.Type, availableType) && mediaType.Subtype == "*":
				if mediaType.Q > bestQ || bestWild > 1 {
					bestQ = mediaType.Q
					bestWild = 1
					bestContentType = available
				}
				continue
			default:
				if strings.EqualFold(mediaType.Type, availableType) && strings.EqualFold(mediaType.Subtype, availableSubtype) &&
					(mediaType.Q > bestQ || bestWild > 0) {
					bestQ = mediaType.Q
					bestWild = 0
					bestContentType = available
				}
			}
		}
	}
	return bestContentType, nil
}

const (
	errNegotiateContentTypeAvailableEmpty              negotiateContentTypeError = "negotiate content type; available content types empty"
	errNegotiateContentTypeInvalidAvailableContentType negotiateContentTypeError = "negotiate content type; invalid available content type"
	errParseAcceptInvalidTag                           negotiateContentTypeError = "parse accept; invalid tag"
	errParseAcceptInvalidTagParameter                  negotiateContentTypeError = "parse accept; invalid tag; invalid parameter"
	errParseAcceptInvalidTagParameterQ                 negotiateContentTypeError = "parse accept; invalid tag; invalid parameter; invalid q"
)

// parseAccept parses the contents of an Accept* header as
// defined in http://www.ietf.org/rfc/rfc2616.txt and returns a list of acceptTags.
//
// The Accept header should be in the form:
//
//	Accept         = "Accept" ":"
//	                 #( media-range [ accept-params ] )
//
//	media-range    = ( "*/*"
//	                 | ( type "/" "*" )
//	                 | ( type "/" subtype )
//	                 ) *( ";" parameter )
//	accept-params  = ";" "q" "=" qvalue *( accept-extension )
//	accept-extension = ";" token [ "=" ( token | quoted-string ) ]
//
// The Tags will be sorted by highest weight first and then by first occurrence.
// Tags with a weight of zero will be dropped. An error will be returned if the
// input could not be parsed.
func parseAccept(input string) (tags []acceptMediaType, err error) {
	if input == "" {
		return
	}
	rawTags := strings.Split(input, ",")
	var parsedTag acceptMediaType
	for _, rawTag := range rawTags {
		parsedTag, err = parseAcceptMediaType(rawTag)
		if err != nil {
			return
		}
		tags = append(tags, parsedTag)
	}
	return
}

type acceptMediaType struct {
	RawMediaType  string
	RawParameters map[string]string

	Type    string
	Subtype string
	Level   int
	Q       float64
}

type negotiateContentTypeError string

func (ncte negotiateContentTypeError) Error() string { return string(ncte) }

func parseAcceptMediaType(rawTag string) (tag acceptMediaType, err error) {
	if rawTag == "" {
		err = errParseAcceptInvalidTag
		return
	}
	tag.RawMediaType, tag.RawParameters, err = mime.ParseMediaType(rawTag)
	if err != nil {
		return
	}
	tag.Q = 1.0 // default?
	tag.Type, tag.Subtype = splitOn(tag.RawMediaType, rune('/'))
	if strings.ContainsRune(tag.Type, rune('*')) && tag.Type != "*" {
		err = errParseAcceptInvalidTag
		return
	}
	if strings.ContainsRune(tag.Subtype, rune('*')) && tag.Subtype != "*" {
		err = errParseAcceptInvalidTag
		return
	}

	for key, value := range tag.RawParameters {
		switch strings.ToLower(key) {
		case "q":
			tag.Q, err = strconv.ParseFloat(value, 32)
			if err != nil {
				return
			}
		case "level":
			tag.Level, err = strconv.Atoi(value)
			if err != nil {
				return
			}
		}
	}

	return
}

func splitOn(s string, split rune) (before, after string) {
	if s == "" {
		return
	}
	runes := []rune(s)
	for index := 0; index < len(runes); index++ {
		if runes[index] == split {
			before = string(runes[:index])
			if index < len(runes)-1 {
				after = string(runes[index+1:])
			}
			return
		}
	}
	before = s
	return
}
