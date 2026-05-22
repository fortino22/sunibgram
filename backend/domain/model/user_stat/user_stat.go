package userstatmodel

type UserStat struct {
	UserID         uint `gorm:"primaryKey;autoIncrement:false"`
	FollowerCount  int  `gorm:"not null;default:0"`
	FollowingCount int  `gorm:"not null;default:0"`
}
