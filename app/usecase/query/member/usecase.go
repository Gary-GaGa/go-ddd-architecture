package query

import (
	"context"
	in "go-ddd-architecture/app/usecase/port/in/query"
)

// GetMemberInfo - 取得會員資訊
func (u *usecase) GetMemberInfo(ctx context.Context, memberNo string) (*in.GetMemberInfoEvent, error) {

	return nil, nil
}
