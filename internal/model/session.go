package model

import "time"

type Session struct {
	UUID      string
	UserUUID  string
	UserLogin string
	ExpiresAt time.Time
}
