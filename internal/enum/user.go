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
