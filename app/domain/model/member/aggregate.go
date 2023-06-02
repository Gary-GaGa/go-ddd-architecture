// Package model -
package model

import "time"

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
	// 會員建立時間
	CreatedAt time.Time
	// 最後登入時間
	LoginedAt time.Time
}
