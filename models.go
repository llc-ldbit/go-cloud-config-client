package configService

import "time"

type ServiceSetting struct {
	Key     string    `json:"key" binding:"required"`
	Value   string    `json:"value" binding:"required"`
	Created time.Time `json:"created" binding:"required"`
	Updated time.Time `json:"updated" binding:"required"`
}

func (s *ServiceSetting) GetValue() string {
	return s.Value
}