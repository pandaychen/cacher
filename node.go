package cacher

import "time"

type UserData struct {
	UKey        string
	UValue      interface{}
	Expired     bool
	ExpireStamp time.Time
	Frequency   uint32 //访问频率
}
