package common

type Role int

const (
	Student   Role = iota + 1 // 1
	Assistant                 // 2
	Teacher                   // 3
	Admin                     // 4
)
