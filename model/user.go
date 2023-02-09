package model

import "time"

type User struct {
	Id         int64      `json:"id,omitempty"`
	Name       string     `json:"name"`
	Surname    string     `json:"surname"`
	Patronymic string     `json:"patronymic,omitempty"`
	Sex        string     `json:"sex"`
	Status     string     `json:"status,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Created    time.Time  `json:"created"`
}

type UserFilter struct {
	Sex        string `json:"sex,omitempty"`
	Status     string `json:"status,omitempty"`
	Name       string `json:"name,omitempty"`
	Surname    string `json:"surname,omitempty"`
	Patronymic string `json:"patronymic,omitempty"`
	OrderBy    string `json:"order_by,omitempty"`
	Desc       *bool  `json:"sort_desc,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
	Offset     uint64 `json:"offset,omitempty"`
}

type ResponseMessage struct {
	Message string `json:"message"`
}

type DeleteUserRequest struct {
	UserID int64 `json:"id"`
}

type ChangeUserRequest struct {
	Id         int64      `json:"id"`
	Name       string     `json:"name,omitempty"`
	Surname    string     `json:"surname,omitempty"`
	Patronymic string     `json:"patronymic,omitempty"`
	Sex        string     `json:"sex,omitempty"`
	Status     string     `json:"status,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
}
