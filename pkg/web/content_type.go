package web

const (
	// ContentTypeApplicationJSON is a content type for JSON responses.
	// We specify chartset=utf-8 so that clients know to use the UTF-8 string encoding.
	ContentTypeApplicationJSON = "application/json; charset=utf-8"

	// ContentTypeApplicationXML is a content type header value.
	ContentTypeApplicationXML = "application/xml"

	// ContentTypeApplicationGob is a content type for Gob responses.
	// We specify chartset=utf-8 so that clients know to use the UTF-8 string encoding.
	ContentTypeApplicationGob = "application/gob"

	// ContentTypeApplicationFormEncoded is a content type header value.
	ContentTypeApplicationFormEncoded = "application/x-www-form-urlencoded"

	// ContentTypeApplicationOctetStream is a content type header value.
	ContentTypeApplicationOctetStream = "application/octet-stream"

	// ContentTypeHTML is a content type for html responses.
	// We specify chartset=utf-8 so that clients know to use the UTF-8 string encoding.
	ContentTypeHTML = "text/html; charset=utf-8"

	//ContentTypeXML is a content type for XML responses.
	// We specify chartset=utf-8 so that clients know to use the UTF-8 string encoding.
	ContentTypeXML = "text/xml; charset=utf-8"

	// ContentTypeText is a content type for text responses.
	// We specify chartset=utf-8 so that clients know to use the UTF-8 string encoding.
	ContentTypeText = "text/plain; charset=utf-8"

	// ContentEncodingIdentity is the identity (uncompressed) content encoding.
	ContentEncodingIdentity = "identity"

	// ContentEncodingGZIP is the gzip (compressed) content encoding.
	ContentEncodingGZIP = "gzip"

	// ConnectionClose is the connection value of "close"
	ConnectionClose = "close"
)
