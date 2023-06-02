// Package model -
package model

// Member -
type Member struct {
	// 會員編號
	No string
	// 會員姓名
	Name string
	// 會員Email
	Email string
	// 會員手機
	Phone *Phone
}
