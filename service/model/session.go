package model

import "time"

// User Session
type (
	// basic session information
	Session struct {
		ID     string
		UserID int64
		AT     string
	}

	// SessionStore is the session data stored in redis
	SessionStore struct {
		Session
		HashedRT   string
		HashedCSRF string
	}

	// SessionOut is the session output for client
	SessionOut struct {
		Session
		RT          string    // plain text
		CSRFToken   string    // plain text
		ATExpiresAt time.Time // access token expires at
	}
)
