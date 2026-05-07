package model

type UserSettings struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false"`
	User      User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	IsPrivate bool `gorm:"not null;default:false"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}
