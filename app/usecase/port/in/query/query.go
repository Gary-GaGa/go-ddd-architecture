// Package in -
package in

import "context"

// QueryUsecase -
type QueryUsecase interface {
	// 取得會員資訊
	GetMemberInfo(ctx context.Context, memberNo string) (*GetMemberInfoEvent, error)
}
