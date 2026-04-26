package web

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// StaticFileServer serves static results for `*filepath` suffix routes.
//
// If you want to use a "cached" mode you should use embedding and http.FS(...) as
// a search path.
type StaticFileServer struct {
	SearchPaths                  []http.FileSystem
	RewriteRules                 []RewriteRule
	Headers                      http.Header
	UseRouteFilepathAsRequestURL bool
	UseEmptyResponseIfNotFound   bool
}

// AddRegexPathRewrite adds a re-write rule that modifies the url path.
//
// Typically these kinds of re-write rules are used for vanity forms or
// removing a cache busting string from a given path.
func (sc *StaticFileServer) AddRegexPathRewrite(match string, rewriteAction func(string, ...string) string) error {
	expr, err := regexp.Compile(match)
	if err != nil {
		return err
	}
	sc.RewriteRules = append(sc.RewriteRules, RewriteRuleFunc(func(path string) (string, bool) {
		if expr.MatchString(path) {
			pieces := flatten(expr.FindAllStringSubmatch(path, -1))
			return rewriteAction(path, pieces...), true
		}
		return path, false
	}))
	return nil
}

// Action implements an action handler.
func (sc StaticFileServer) Action(ctx Context) Result {
	// we _must_ do this to ensure that
	// file paths will match when we look for them
	// in the static asset path(s).
	path, ok := ctx.RouteParams().Get("filepath")
	if ok {
		ctx.Request().URL.Path = path
	}
	sc.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

// ServeHTTP is the entrypoint for the static server.
//
// It  adds default headers if specified, and then serves the file from disk.
func (sc StaticFileServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for key, values := range sc.Headers {
		for _, value := range values {
			rw.Header().Set(key, value)
		}
	}
	filePath := req.URL.Path
	f, finalPath, err := sc.ResolveFile(filePath)
	if err != nil {
		sc.fileError(finalPath, rw, req, err)
		return
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		sc.fileError(finalPath, rw, req, err)
		return
	}
	if finfo.IsDir() {
		sc.notFound(finalPath, rw, req)
		return
	}
	http.ServeContent(rw, req, finalPath, finfo.ModTime(), f)
}

// ResolveFile resolves a file from rewrite rules and search paths.
//
// First the file path is modified according to the rewrite rules.
// Then each search path is checked for the resolved file path.
func (sc StaticFileServer) ResolveFile(filePath string) (f http.File, finalPath string, err error) {
	for _, rule := range sc.RewriteRules {
		if newFilePath, matched := rule.Apply(filePath); matched {
			filePath = newFilePath
		}
	}
	for _, searchPath := range sc.SearchPaths {
		f, err = searchPath.Open(filePath)
		if typed, ok := f.(fileNamer); ok && typed != nil {
			finalPath = typed.Name()
		} else {
			finalPath = filePath
		}
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return
		}
		if f != nil {
			return
		}
	}
	return
}

func (sc StaticFileServer) fileError(name string, rw http.ResponseWriter, req *http.Request, err error) {
	if os.IsNotExist(err) {
		sc.notFound(name, rw, req)
		return
	}
	http.Error(rw, err.Error(), http.StatusInternalServerError)
}

func (sc StaticFileServer) notFound(name string, rw http.ResponseWriter, req *http.Request) {
	if sc.UseEmptyResponseIfNotFound {
		ctype := mime.TypeByExtension(filepath.Ext(name))
		rw.Header().Set(HeaderContentType, ctype)
		rw.Header().Set(HeaderContentLength, "0")
		rw.WriteHeader(http.StatusOK)
		return
	}
	http.NotFound(rw, req)
}

// RewriteRule is a type that modifies a request url path.
//
// It should take the `url.Path` value, and return an updated
// value and true, or the value unchanged and false.
type RewriteRule interface {
	Apply(string) (string, bool)
}

// RewriteRuleFunc is a function that rewrites a url.
type RewriteRuleFunc func(string) (string, bool)

// Apply applies the rewrite rule.
func (rrf RewriteRuleFunc) Apply(path string) (string, bool) {
	return rrf(path)
}

func flatten(pieces [][]string) (output []string) {
	for _, p := range pieces {
		output = append(output, p...)
	}
	return
}

type fileNamer interface {
	Name() string
}
