package entity

import "time"

type Stats struct {
	Level     int       `json:"level"`
	MaxHP     int       `json:"maxHp"`
	HP        int       `json:"hp"`
	MaxMP     int       `json:"maxMp"`
	MP        int       `json:"mp"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}