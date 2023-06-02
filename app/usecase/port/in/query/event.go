package in

import (
	dto "go-ddd-architecture/app/usecase/dto/query"
)

// GetMemberInfoEvent -
type GetMemberInfoEvent struct {
	// 會員編號
	No string
	// 會員姓名
	Name string
	// 會員Email
	Email string
	// 會員手機
	Phone *dto.PhoneDto
}
