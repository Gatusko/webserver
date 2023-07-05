package structs

type User struct {
	Id    int    `json:"id,omitempty"`
	Email string `json:"email"`
}
