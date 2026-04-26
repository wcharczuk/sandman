package web

import "encoding/json"

// BodyAsJSON reads a request body and deserializes
// it as json into a given reference.
func BodyAsJSON(r Context, ref any) error {
	return json.NewDecoder(r.Request().Body).Decode(ref)
}
