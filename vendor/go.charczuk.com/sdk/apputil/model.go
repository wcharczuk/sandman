package apputil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
)

var (
	userCols          = db.TypeMetaFor(User{})
	userTableName     = User{}.TableName()
	sessionsTableName = Session{}.TableName()

	getUserByEmailQuery        = fmt.Sprintf("SELECT %s FROM %s WHERE email = $1", db.ColumnNamesCSV(userCols.Columns()), userTableName)
	deleteExpiredSessionsQuery = fmt.Sprintf("DELETE FROM %s WHERE expires_utc < $1", sessionsTableName)
)

// NewModelManager returns a new model manager.
func NewModelManager(conn *db.Connection, opts ...db.InvocationOption) *ModelManager {
	return &ModelManager{
		BaseManager: dbutil.NewBaseManager(conn, opts...),
	}
}

// ModelManager implements database functions.
type ModelManager struct {
	dbutil.BaseManager
}

// GetUserByEmail gets a user by email.
func (m ModelManager) GetUserByEmail(ctx context.Context, email string) (output User, found bool, err error) {
	found, err = m.Invoke(ctx, db.OptLabel("get_user_by_email")).Query(getUserByEmailQuery, email).Out(&output)
	return
}

// DeleteExpiredSessions deletes sessions that are expired and older than 14 days.
func (m ModelManager) DeleteExpiredSessions(ctx context.Context, oldest time.Time) error {
	_, err := m.Invoke(ctx, db.OptLabel("delete_expired_sessions")).Exec(deleteExpiredSessionsQuery, oldest)
	return err
}

// NewTest returns a new test context.
func NewTest(t *testing.T) (*ModelManager, func()) {
	tx, err := testutil.DefaultDB().BeginTx(context.Background())
	assert.ItsNil(t, err)
	return NewModelManager(testutil.DefaultDB(), db.OptTx(tx)), func() {
		_ = tx.Rollback()
	}
}

// CreateTestUser creates a test user.
func CreateTestUser(t *testing.T, mgr *ModelManager) *User {
	u0 := NewTestUser()
	assert.ItsNil(t, mgr.Invoke(context.Background()).Create(&u0))
	return &u0
}

// CreateTestSession creates a test session.
func CreateTestSession(t *testing.T, mgr *ModelManager, user *User) *Session {
	s0 := NewTestSession(user)
	assert.ItsNil(t, mgr.Invoke(context.Background()).Create(&s0))
	return &s0
}
