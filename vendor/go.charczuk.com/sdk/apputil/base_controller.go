package apputil

import (
	"go.charczuk.com/sdk/uuid"
	"go.charczuk.com/sdk/web"
)

// BaseController holds useful common methods for controllers.
type BaseController struct{}

// GetUserID gets a userID from a given web context.
//
// You can use this as a proxy for determining if a
// session is authenticated or not.
func (bc BaseController) GetUserID(r web.Context) uuid.UUID {
	return uuid.MustParse(r.Session().UserID)
}
