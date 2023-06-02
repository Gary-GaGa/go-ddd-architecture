package po

// PhonePo -
type PhonePo struct {
	// 國碼
	CountryCode string `bson:"countryCode"`
	// 手機號碼
	PhoneNumber string `bson:"phoneNumber"`
}
