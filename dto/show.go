package dto

import "github.com/haschi/dinge/model"

type ShowResponseData struct {
	model.Ding
	History []model.Event
}
