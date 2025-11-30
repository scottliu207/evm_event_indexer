package enum

type UserStatus int8

const (
	_ UserStatus = iota
	UserStatusEnabled
	UserStatusDisabled
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusEnabled:
		return "enabled"
	case UserStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

type UserRole int8

const (
	_ UserRole = iota
	UserRoleAdmin
	UserRoleUser
)

func (r UserRole) String() string {
	switch r {
	case UserRoleAdmin:
		return "admin"
	case UserRoleUser:
		return "user"
	default:
		return "unknown"
	}
}
