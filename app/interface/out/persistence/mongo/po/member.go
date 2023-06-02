// Package po -
package po

import "time"

// MemberPo -
type MemberPo struct {
	// 會員編號
	No string `bson:"no"`
	// 會員姓名
	Name string `bson:"name"`
	// 會員Email
	Email string `bson:"email"`
	// 會員手機
	Phone *PhonePo `bson:"phone"`
	// 會員建立時間
	CreatedAt time.Time `bson:"createdAt"`
	// 最後登入時間
	LoginedAt time.Time `bson:"loginedAt"`
	// 更新時間
	UpdatedAt time.Time `bson:"updatedAt"`
}
