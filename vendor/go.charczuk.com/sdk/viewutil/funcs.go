package viewutil

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"go.charczuk.com/sdk/mathutil"
	"go.charczuk.com/sdk/stringutil"
	"go.charczuk.com/sdk/uuid"
)

// Funcs is a singleton for viewfuncs.
var (
	Funcs template.FuncMap = map[string]any{
		"as_bytes":                   AsBytes,
		"as_duration":                AsDuration,
		"as_string":                  AsString,
		"at_index":                   AtIndex,
		"base64":                     Base64,
		"base64decode":               Base64Decode,
		"ceil":                       Ceil,
		"concat":                     Concat,
		"contains":                   Contains,
		"control_for":                ControlFor,
		"csv":                        CSV,
		"duration_round_millis":      DurationRoundMillis,
		"duration_round_seconds":     DurationRoundSeconds,
		"duration_round":             DurationRound,
		"first":                      First,
		"floor":                      Floor,
		"format_filesize":            FormatFileSize,
		"format_money":               FormatMoney,
		"format_pct":                 FormatPct,
		"form_for":                   FormFor,
		"generate_ordinal_names":     GenerateOrdinalNames,
		"has_prefix":                 HasPrefix,
		"has_suffix":                 HasSuffix,
		"indent_spaces":              IndentSpaces,
		"indent_tabs":                IndentTabs,
		"join":                       Join,
		"last":                       Last,
		"matches":                    Matches,
		"parse_bool":                 ParseBool,
		"parse_float64":              ParseFloat64,
		"parse_int":                  ParseInt,
		"parse_int64":                ParseInt64,
		"parse_json":                 ParseJSON,
		"parse_time_unix":            ParseTimeUnix,
		"parse_time":                 ParseTime,
		"parse_url":                  ParseURL,
		"parse_uuid":                 ParseUUID,
		"prefix":                     Prefix,
		"quote":                      Quote,
		"reverse":                    Reverse,
		"round":                      Round,
		"sequence_range":             SequenceRange,
		"sha256":                     SHA256,
		"sha512":                     SHA512,
		"slice":                      Slice,
		"slugify":                    Slugify,
		"split_n":                    SplitN,
		"split":                      Split,
		"strip_quotes":               StripQuotes,
		"suffix":                     Suffix,
		"time_day":                   TimeDay,
		"time_format_date_long":      TimeFormatDateLong,
		"time_format_date_month_day": TimeFormatDateMonthDay,
		"time_format_date_short_rev": TimeFormatDateShortRev,
		"time_format_date_short":     TimeFormatDateShort,
		"time_format_kitchen":        TimeFormatKitchen,
		"time_format_medium":         TimeFormatMedium,
		"time_format_rfc3339":        TimeFormatRFC3339,
		"time_format_short":          TimeFormatShort,
		"time_format":                TimeFormat,
		"time_hour":                  TimeHour,
		"time_in_loc":                TimeInLocation,
		"time_in_utc":                TimeInUTC,
		"time_is_epoch":              TimeIsEpoch,
		"time_is_zero":               TimeIsZero,
		"time_millisecond":           TimeMillisecond,
		"time_minute":                TimeMinute,
		"time_month":                 TimeMonth,
		"time_now_utc":               TimeNowUTC,
		"time_now":                   TimeNow,
		"time_second":                TimeSecond,
		"time_since_utc":             TimeSinceUTC,
		"time_since":                 TimeSince,
		"time_sub":                   TimeSub,
		"time_unix_nano":             TimeUnixNano,
		"time_unix":                  TimeUnix,
		"time_year":                  TimeYear,
		"to_json_pretty":             ToJSONPretty,
		"to_json":                    ToJSON,
		"to_lower":                   ToLower,
		"to_title":                   ToTitle,
		"to_upper":                   ToUpper,
		"trim_prefix":                TrimPrefix,
		"trim_space":                 TrimSpace,
		"trim_suffix":                TrimSuffix,
		"tsv":                        TSV,
		"url_host":                   URLHost,
		"url_path":                   URLPath,
		"url_port":                   URLPort,
		"url_query":                  URLQuery,
		"url_raw_query":              URLRawQuery,
		"url_scheme":                 URLScheme,
		"urlencode":                  URLEncode,
		"uuid":                       UUIDv4,
		"uuidv4":                     UUIDv4,
		"with_url_host":              WithURLHost,
		"with_url_path":              URLPath,
		"with_url_port":              WithURLPort,
		"with_url_query":             WithURLQuery,
		"with_url_raw_query":         WithURLRawQuery,
		"with_url_scheme":            WithURLScheme,
	}
)

// FileExists returns if the file at a given path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads the contents of a file path as a string.
func ReadFile(path string) (string, error) {
	contents, err := os.ReadFile(path)
	return string(contents), err
}

// AsString attempts to return a string representation of a value.
func AsString(v any) string {
	switch c := v.(type) {
	case []byte:
		return string(c)
	case string:
		return c
	default:
		return fmt.Sprintf("%v", v)
	}
}

// AsBytes attempts to return a bytes representation of a value.
func AsBytes(v interface{}) []byte {
	return []byte(fmt.Sprintf("%v", v))
}

// ParseInt parses a value as an integer.
func ParseInt(v interface{}) (int, error) {
	return strconv.Atoi(fmt.Sprintf("%v", v))
}

// ParseInt64 parses a value as an int64.
func ParseInt64(v interface{}) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
}

// ParseFloat64 parses a value as a float64.
func ParseFloat64(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

// ParseBool attempts to parse a value as a bool.
// "truthy" values include "true", "1", "yes".
// "falsey" values include "false", "0", "no".
func ParseBool(raw interface{}) (bool, error) {
	v := fmt.Sprintf("%v", raw)
	if len(v) == 0 {
		return false, nil
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value `%s`", v)
	}
}

// ParseTime parses a time string with a given format.
func ParseTime(format, v string) (time.Time, error) {
	return time.Parse(format, v)
}

// ParseTimeUnix returns a timestamp from a unix format.
func ParseTimeUnix(v int64) time.Time {
	return time.Unix(v, 0)
}

// TimeNow returns the current time in the system timezone.
func TimeNow() time.Time {
	return time.Now()
}

// TimeNowUTC returns the current time in the UTC timezone.
func TimeNowUTC() time.Time {
	return time.Now().UTC()
}

// TimeUnix returns the unix format for a timestamp.
func TimeUnix(t time.Time) int64 {
	return t.Unix()
}

// TimeUnixNano returns the timetamp as nanoseconds from 1970-01-01.
func TimeUnixNano(t time.Time) int64 {
	return t.UnixNano()
}

// TimeFormatRFC3339 returns the RFC3339 format for a timestamp.
func TimeFormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// TimeFormatShort returns the short format for a timestamp.
// The format string is "1/02/2006 3:04:05 PM".
func TimeFormatShort(t time.Time) string {
	return t.Format("1/02/2006 3:04:05 PM")
}

// TimeFormat returns the time with a given format string.
func TimeFormat(format string, t time.Time) string {
	return t.Format(format)
}

// TimeIsZero returns if the time is set or not.
func TimeIsZero(t time.Time) bool {
	return t.IsZero()
}

// TimeIsEpoch returns if the time is the unix epoch time or not.
func TimeIsEpoch(t time.Time) bool {
	return t.Equal(time.Unix(0, 0))
}

// TimeFormatDateLong returns the short date for a timestamp.
func TimeFormatDateLong(t time.Time) string {
	return t.Format("Jan _2, 2006")
}

// TimeFormatDateShort returns the short date for a timestamp.
// The format string is "1/02/2006"
func TimeFormatDateShort(t time.Time) string {
	return t.Format("1/02/2006")
}

// TimeFormatDateShortRev returns the short date for a timestamp in YYYY/mm/dd format.
func TimeFormatDateShortRev(t time.Time) string {
	return t.Format("2006/1/02")
}

// TimeFormatTimeMedium returns the medium format for a timestamp.
// The format string is "1/02/2006 3:04:05 PM".
func TimeFormatMedium(t time.Time) string {
	return t.Format("Jan 02, 2006 3:04:05 PM")
}

// TimeFormatTimeKitchen returns the kitchen format for a timestamp.
// The format string is "3:04PM".
func TimeFormatKitchen(t time.Time) string {
	return t.Format(time.Kitchen)
}

// TimeFormatDateMonthDay returns the month dat format for a timestamp.
// The format string is "1/2".
func TimeFormatDateMonthDay(t time.Time) string {
	return t.Format("1/2")
}

// TimeInUTC returns the time in a given location by string.
// If the location is invalid, this will error.
func TimeInUTC(t time.Time) time.Time {
	return t.UTC()
}

// TimeInLocation returns the time in a given location by string.
// If the location is invalid, this will error.
func TimeInLocation(loc string, t time.Time) (time.Time, error) {
	location, err := time.LoadLocation(loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(location), err
}

// TimeYear returns the year component of a timestamp.
func TimeYear(t time.Time) int {
	return t.Year()
}

// TimeMonth returns the month component of a timestamp.
func TimeMonth(t time.Time) int {
	return int(t.Month())
}

// TimeDay returns the day component of a timestamp.
func TimeDay(t time.Time) int {
	return t.Day()
}

// TimeHour returns the hour component of a timestamp.
func TimeHour(t time.Time) int {
	return t.Hour()
}

// TimeMinute returns the minute component of a timestamp.
func TimeMinute(t time.Time) int {
	return t.Minute()
}

// TimeSecond returns the seconds component of a timestamp.
func TimeSecond(t time.Time) int {
	return t.Second()
}

// TimeMillisecond returns the millisecond component of a timestamp.
func TimeMillisecond(t time.Time) int {
	return int(time.Duration(t.Nanosecond()) / time.Millisecond)
}

// AsDuration returns a given value as a duration.
func AsDuration(val any) (typedVal time.Duration, err error) {
	switch tv := val.(type) {
	case time.Duration:
		typedVal = tv
	case uint8:
		typedVal = time.Duration(tv)
	case int8:
		typedVal = time.Duration(tv)
	case uint16:
		typedVal = time.Duration(tv)
	case int16:
		typedVal = time.Duration(tv)
	case uint32:
		typedVal = time.Duration(tv)
	case int32:
		typedVal = time.Duration(tv)
	case uint64:
		typedVal = time.Duration(tv)
	case int64:
		typedVal = time.Duration(tv)
	case int:
		typedVal = time.Duration(tv)
	case uint:
		typedVal = time.Duration(tv)
	case float32:
		typedVal = time.Duration(tv)
	case float64:
		typedVal = time.Duration(tv)
	default:
		err = fmt.Errorf("invalid duration value %[1]T: %[1]v", val)
	}
	return
}

// TimeSince returns the duration since a given timestamp.
// It is relative, meaning the value returned can be negative.
func TimeSince(t time.Time) time.Duration {
	return time.Since(t)
}

// TimeSub the duration difference between two times.
func TimeSub(t1, t2 time.Time) time.Duration {
	return t1.UTC().Sub(t2.UTC())
}

// TimeSinceUTC returns the duration since a given timestamp in UTC.
// It is relative, meaning the value returned can be negative.
func TimeSinceUTC(t time.Time) time.Duration {
	return time.Now().UTC().Sub(t.UTC())
}

// DurationRound rounds a duration value.
func DurationRound(d time.Duration, to time.Duration) time.Duration {
	return d.Round(to)
}

// DurationRoundMillis rounds a duration value to milliseconds.
func DurationRoundMillis(d time.Duration) time.Duration {
	return d.Round(time.Millisecond)
}

// DurationRoundSeconds rounds a duration value to seconds.
func DurationRoundSeconds(d time.Duration) time.Duration {
	return d.Round(time.Millisecond)
}

// Round returns the value rounded to a given set of places.
// It uses midpoint rounding.
func Round(places, d float64) float64 {
	return mathutil.RoundPlaces(d, int(places))
}

// Ceil returns the value rounded up to the nearest integer.
func Ceil(d float64) float64 {
	return math.Ceil(d)
}

// Floor returns the value rounded down to zero.
func Floor(d float64) float64 {
	return math.Floor(d)
}

// FormatMoney returns a float as a formatted string rounded to two decimal places.
func FormatMoney(d float64) string {
	return fmt.Sprintf("$%0.2f", mathutil.RoundPlaces(d, 2))
}

// FormatPct formats a float as a percentage (it is multiplied by 100,
// then suffixed with '%')
func FormatPct(d float64) string {
	return fmt.Sprintf("%0.2f%%", d*100)
}

// FormatFileSize formats an int as a file size.
func FormatFileSize(sizeBytes int) string {
	if sizeBytes >= 1<<30 {
		return fmt.Sprintf("%dgB", sizeBytes/(1<<30))
	} else if sizeBytes >= 1<<20 {
		return fmt.Sprintf("%dmB", sizeBytes/(1<<20))
	} else if sizeBytes >= 1<<10 {
		return fmt.Sprintf("%dkB", sizeBytes/(1<<10))
	}
	return fmt.Sprintf("%dB", sizeBytes)
}

// Base64 encodes data as a string as a base6 string.
func Base64(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

// Base64Decode decodes a base 64 string.
func Base64Decode(v string) (string, error) {
	result, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ParseUUID parses a uuid.
func ParseUUID(v string) (uuid.UUID, error) {
	return uuid.Parse(v)
}

// UUIDv4 generates a uuid v4.
func UUIDv4() uuid.UUID {
	return uuid.V4()
}

// ToUpper returns a string case shifted to upper case.
func ToUpper(v string) string {
	return strings.ToUpper(v)
}

// ToLower returns a string case shifted to lower case.
func ToLower(v string) string {
	return strings.ToLower(v)
}

// ToTitle returns a title cased string.
func ToTitle(v string) string {
	return strings.ToTitle(v)
}

// Slugify returns a slug format string.
// It replaces whitespace with `-`
// It path escapes any other characters.
func Slugify(v string) string {
	return stringutil.Slugify(v)
}

// TrimSpace trims whitespace from the beginning and end of a string.
func TrimSpace(v string) string {
	return strings.TrimSpace(v)
}

// Prefix appends a given string to a prefix.
func Prefix(pref, v string) string {
	return pref + v
}

// Concat concatenates a list of strings.
func Concat(strs ...string) string {
	var output string
	for index := 0; index < len(strs); index++ {
		output = output + strs[index]
	}
	return output
}

// Suffix appends a given prefix to a string.
func Suffix(suf, v string) string {
	return v + suf
}

// Split splits a string by a separator.
func Split(sep, v string) []string {
	return strings.Split(v, sep)
}

// SplitN splits a string by a separator a given number of times.
func SplitN(sep string, n float64, v string) []string {
	return strings.SplitN(v, sep, int(n))
}

//
// array functions
//

// Reverse reverses an array.
func Reverse(collection interface{}) (interface{}, error) {
	value := reflect.ValueOf(collection)

	if value.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice")
	}

	output := make([]interface{}, value.Len())
	for index := 0; index < value.Len(); index++ {
		output[index] = value.Index((value.Len() - 1) - index).Interface()
	}
	return output, nil
}

// Slice returns a subrange of a collection.
func Slice(from, to int, collection interface{}) (interface{}, error) {
	value := reflect.ValueOf(collection)

	if value.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice")
	}

	return value.Slice(from, to).Interface(), nil
}

// First returns the first element of a collection.
func First(collection interface{}) (interface{}, error) {
	value := reflect.ValueOf(collection)
	kind := value.Type().Kind()
	if kind != reflect.Slice && kind != reflect.Map && kind != reflect.Array {
		return nil, fmt.Errorf("input must be a slice or map")
	}
	if value.Len() == 0 {
		return nil, nil
	}
	switch kind {
	case reflect.Slice, reflect.Array:
		return value.Index(0).Interface(), nil
	case reflect.Map:
		iter := value.MapRange()
		if iter.Next() {
			return iter.Value().Interface(), nil
		}
	default:
	}

	return nil, nil
}

// AtIndex returns an element at a given index.
func AtIndex(index int, collection interface{}) (interface{}, error) {
	value := reflect.ValueOf(collection)
	if value.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice")
	}
	if value.Len() == 0 {
		return nil, nil
	}
	return value.Index(index).Interface(), nil
}

// Last returns the last element of a collection.
func Last(collection interface{}) (interface{}, error) {
	value := reflect.ValueOf(collection)
	if value.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice")
	}
	if value.Len() == 0 {
		return nil, nil
	}
	return value.Index(value.Len() - 1).Interface(), nil
}

// Join creates a string joined with a given separator.
func Join(sep string, collection any) (string, error) {
	value := reflect.ValueOf(collection)
	if value.Type().Kind() != reflect.Slice {
		return "", fmt.Errorf("input must be a slice")
	}
	if value.Len() == 0 {
		return "", nil
	}
	values := make([]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		values[i] = fmt.Sprintf("%v", value.Index(i).Interface())
	}
	return strings.Join(values, sep), nil
}

// CSV returns a csv of a given collection.
func CSV(collection interface{}) (string, error) {
	return Join(",", collection)
}

// TSV returns a tab separated values of a given collection.
func TSV(collection interface{}) (string, error) {
	return Join("\t", collection)
}

// HasSuffix returns if a string has a given suffix.
func HasSuffix(suffix, v string) bool {
	return strings.HasSuffix(v, suffix)
}

// HasPrefix returns if a string has a given prefix.
func HasPrefix(prefix, v string) bool {
	return strings.HasPrefix(v, prefix)
}

// TrimSuffix returns if a string has a given suffix.
func TrimSuffix(suffix, v string) string {
	return strings.TrimSuffix(v, suffix)
}

// TrimPrefix returns if a string has a given prefix.
func TrimPrefix(prefix, v string) string {
	return strings.TrimPrefix(v, prefix)
}

// Contains returns if a string contains a given substring.
func Contains(substr, v string) bool {
	return strings.Contains(v, substr)
}

// Matches returns if a string matches a given regular expression.
func Matches(expr, v string) (bool, error) {
	return regexp.MatchString(expr, v)
}

// Quote returns a string wrapped in " characters.
// It will trim space before and after, and only add quotes
// if they don't already exist.
func Quote(v string) string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "\"") {
		v = "\"" + v
	}
	if !strings.HasSuffix(v, "\"") {
		v = v + "\""
	}
	return v
}

// StripQuotes strips leading and trailing quotes.
func StripQuotes(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "\"")
	v = strings.TrimSuffix(v, "\"")
	return v
}

// ParseURL parses a url.
func ParseURL(v string) (*url.URL, error) {
	return url.Parse(v)
}

// URLEncode encodes a value as a url token.
func URLEncode(value string) string {
	return url.QueryEscape(value)
}

// URLScheme returns the scheme of a url.
func URLScheme(v *url.URL) string {
	return v.Scheme
}

// WithURLScheme returns the scheme of a url.
func WithURLScheme(scheme string, u *url.URL) *url.URL {
	copy := *u
	copy.Scheme = scheme
	return &copy
}

// URLHost returns the host of a url.
func URLHost(v *url.URL) string {
	return v.Host
}

// WithURLHost returns the host of a url.
func WithURLHost(host string, u *url.URL) *url.URL {
	copy := *u
	copy.Host = host
	return &copy
}

// URLPort returns the url port.
// If none is explicitly specified, this will return empty string.
func URLPort(v *url.URL) string {
	return v.Port()
}

// WithURLPort sets the url port.
func WithURLPort(port string, u *url.URL) *url.URL {
	copy := *u
	host := copy.Host
	if strings.Contains(host, ":") {
		parts := strings.SplitN(host, ":", 2)
		copy.Host = parts[0] + ":" + port
	} else {
		copy.Host = host + ":" + port
	}
	return &copy
}

// URLPath returns the url path.
func URLPath(v *url.URL) string {
	return v.Path
}

// WithURLPath returns the url path.
func WithURLPath(path string, u *url.URL) *url.URL {
	copy := *u
	copy.Path = path
	return &copy
}

// URLRawQuery returns the url raw query.
func URLRawQuery(v *url.URL) string {
	return v.RawQuery
}

// WithURLRawQuery returns the url path.
func WithURLRawQuery(rawQuery string, u *url.URL) *url.URL {
	copy := *u
	copy.RawQuery = rawQuery
	return &copy
}

// URLQuery returns a url query param.
func URLQuery(name string, v *url.URL) string {
	return v.Query().Get(name)
}

// WithURLQuery returns a url query param.
func WithURLQuery(key, value string, u *url.URL) *url.URL {
	copy := *u
	queryValues := copy.Query()
	queryValues.Add(key, value)
	copy.RawQuery = queryValues.Encode()
	return &copy
}

// SHA256 returns the sha256 sum of a string.
func SHA256(v string) string {
	h := sha256.New()
	fmt.Fprint(h, v)
	return hex.EncodeToString(h.Sum(nil))
}

// SHA512 returns the sha512 sum of a string.
func SHA512(v string) string {
	h := sha512.New()
	fmt.Fprint(h, v)
	return hex.EncodeToString(h.Sum(nil))
}

// HMAC512 returns the hmac signed sha 512 sum of a string.
func HMAC512(key, v string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha512.New, keyBytes)
	fmt.Fprint(h, v)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// IndentTabs indents a string with a given number of tabs.
func IndentTabs(tabCount int, v interface{}) string {
	lines := strings.Split(fmt.Sprintf("%v", v), "\n")
	outputLines := make([]string, len(lines))

	var tabs string
	for i := 0; i < tabCount; i++ {
		tabs = tabs + "\t"
	}

	for i := 0; i < len(lines); i++ {
		outputLines[i] = tabs + lines[i]
	}
	return strings.Join(outputLines, "\n")
}

// IndentSpaces indents a string by a given set of spaces.
func IndentSpaces(spaceCount int, v interface{}) string {
	lines := strings.Split(fmt.Sprintf("%v", v), "\n")
	outputLines := make([]string, len(lines))

	var spaces string
	for i := 0; i < spaceCount; i++ {
		spaces = spaces + " "
	}

	for i := 0; i < len(lines); i++ {
		outputLines[i] = spaces + lines[i]
	}
	return strings.Join(outputLines, "\n")
}

// GenerateOrdinalNames generates ordinal names by passing the index to a given formatter.
// The formatter should be in Sprintf format (i.e. using a '%d' token for where the index should go).
/*
Example:
    {{ generate_ordinal_names "worker-%d" 3 }} // [worker-0 worker-1 worker-2]
*/
func GenerateOrdinalNames(format string, replicas int) []string {
	output := make([]string, replicas)
	for index := 0; index < replicas; index++ {
		output[index] = fmt.Sprintf(format, index)
	}
	return output
}

// GenerateKey generates a key of a given size base 64 encoded.
func GenerateKey(keySize int) string {
	key := make([]byte, keySize)
	_, _ = io.ReadFull(rand.Reader, key)
	return base64.StdEncoding.EncodeToString(key)
}

// ToJSON returns an object encoded as json.
func ToJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}

// ToJSONPretty encodes an object as json with indentation.
func ToJSONPretty(v any) (string, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseJSON returns an object encoded as json.
func ParseJSON(v string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(v), &data)
	return data, err
}

// SequenceRange returns an array of ints from min to max, not including max.
// Given (0,5) as inputs, it would return [0,1,2,3,4]
func SequenceRange(start, end int) []int {
	if start == end {
		return []int{}
	}
	if start > end {
		output := make([]int, start-end)
		for x := start; x > end; x-- {
			output[start-x] = x
		}
		return output
	}

	output := make([]int, end-start)
	for x := start; x < end; x++ {
		output[x] = x
	}
	return output
}

// AsFloat64 returns a given value as a float64.
func AsFloat64(val interface{}) (typedVal float64, err error) {
	switch tv := val.(type) {
	case uint8:
		typedVal = float64(tv)
	case int8:
		typedVal = float64(tv)
	case uint16:
		typedVal = float64(tv)
	case int16:
		typedVal = float64(tv)
	case uint32:
		typedVal = float64(tv)
	case int32:
		typedVal = float64(tv)
	case uint64:
		typedVal = float64(tv)
	case int64:
		typedVal = float64(tv)
	case int:
		typedVal = float64(tv)
	case float32:
		typedVal = float64(tv)
	case float64:
		typedVal = tv
	default:
		err = fmt.Errorf("invalid to_float value %[1]T: %[1]v", val)
	}
	return
}

// AsInt returns a given value as a int64.
func AsInt(val interface{}) (typedVal int, err error) {
	switch tv := val.(type) {
	case uint8:
		typedVal = int(tv)
	case int8:
		typedVal = int(tv)
	case uint16:
		typedVal = int(tv)
	case int16:
		typedVal = int(tv)
	case uint32:
		typedVal = int(tv)
	case int32:
		typedVal = int(tv)
	case uint64:
		typedVal = int(tv)
	case int64:
		typedVal = int(tv)
	case int:
		typedVal = tv
	case float32:
		typedVal = int(tv)
	case float64:
		typedVal = int(tv)
	default:
		err = fmt.Errorf("invalid to_int value %[1]T: %[1]v", val)
	}
	return
}
