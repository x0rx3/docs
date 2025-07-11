package dto

import "time"

type Meta struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	File     bool      `json:"file"`
	Public   bool      `json:"public"`
	Token    string    `json:"token,omitempty"`
	CreateAt time.Time `json:"create_at,omitempty"`
	Mime     string    `json:"mime"`
	Grant    []string  `json:"grant"`
}
