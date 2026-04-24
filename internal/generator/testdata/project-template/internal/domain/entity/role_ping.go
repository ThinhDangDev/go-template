package entity

type RolePing struct {
	Message string `json:"message"`
	User    *User  `json:"user"`
}
