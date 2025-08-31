package testutils

type UserIdentity struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type DummyEvent struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	UserIdentity *UserIdentity `json:"user_identity,omitempty"`
}
