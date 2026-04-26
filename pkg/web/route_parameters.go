package web

// RouteParameters are parameters sourced from parsing the request path (route).
type RouteParameters []RouteParameter

// RouteParameter is a parameter key and value for the route.
type RouteParameter struct {
	Key, Value string
}

// Get gets a value for a key.
func (rp RouteParameters) Get(key string) (string, bool) {
	for _, kv := range rp {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return "", false
}
