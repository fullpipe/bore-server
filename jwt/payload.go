package jwt

type Payload struct {
	UserID uint
	Roles  []string
}
