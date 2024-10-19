package web

import "net/http"

// Static serves a static file from a given filesystem.
//
// It assumes views as the provider for failure conditions.
func Static(ctx Context, fs http.FileSystem, path string) Result {
	f, err := fs.Open(path)
	if err != nil {
		return ctx.Views().InternalError(err)
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		return ctx.Views().InternalError(err)
	}
	if finfo.IsDir() {
		return ctx.Views().NotFound()
	}
	http.ServeContent(ctx.Response(), ctx.Request(), path, finfo.ModTime(), f)
	return nil
}
