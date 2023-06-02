// Package convert -
package convert

import (
	model "go-ddd-architecture/app/domain/model/member"
	po "go-ddd-architecture/app/interface/out/persistence/mongo/po"
)

// MemberToModel -
func MemberToModel(in *po.MemberPo) *model.Member {
	if in == nil {
		return new(model.Member)
	}

	return &model.Member{
		No:        in.No,
		Name:      in.Name,
		Email:     in.Email,
		Phone:     PhoneToModel(in.Phone),
		CreatedAt: in.CreatedAt,
		LoginedAt: in.LoginedAt,
	}
}

// PhoneToModel -
func PhoneToModel(in *po.PhonePo) *model.Phone {
	if in == nil {
		return new(model.Phone)
	}

	return &model.Phone{
		CountryCode: in.CountryCode,
		PhoneNumber: in.PhoneNumber,
	}
}
