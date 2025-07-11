package dto

type User struct {
	UUID     string `json:"uuid"`
	Login    string `json:"login"`
	Password string `json:"pswd"`
}
