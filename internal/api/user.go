package api

type User struct {
	DeviceToken string `json:"deviceToken"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	ID          string `json:"id"`
	LastName    string `json:"last_name"`
	UniqueID    string `json:"unique_id"`
	Username    string `json:"username"`
}
