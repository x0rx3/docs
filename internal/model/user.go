package model

type User struct {
	UUID     string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:uuid"`
	Login    string `gom:"type:text;not null;cloumn:login"`
	Password string `gorm:"type:text;not null;column:password"`
}

func (inst User) TableName() string {
	return "users"
}
