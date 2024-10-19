package apputil

import (
	"strings"

	"go.charczuk.com/sdk/web"
)

// UpgradeHTTPS upgrades http connections to https with a redirect in typical settings.
//
// We do so through `X-Forwarded-Proto` header sniffs because many reverse-proxy hosted services
// forward all traffic to http and add forwarded headers so we know what the context
// of the original request was.
//
// Normally we'd run a second http server on 80 (in addition to 443) and do this redirect,
// but because of the specific nature of how many systems host "https" services on http ports
// this is what we need to do.
func UpgradeHTTPS(action web.Action) web.Action {
	return func(ctx web.Context) web.Result {
		if strings.EqualFold(ctx.Request().Header.Get(web.HeaderXForwardedProto), "http") {
			web.RedirectUpgrade(ctx.Response(), ctx.Request())
			return nil
		}
		return action(ctx)
	}
}
