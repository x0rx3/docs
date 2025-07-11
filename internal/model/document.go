package model

import "time"

type Document struct {
	UUID     string
	Name     string
	Mime     string
	File     bool
	Public   bool
	CreateAt time.Time
	Grant    []string
	Path     string
}
