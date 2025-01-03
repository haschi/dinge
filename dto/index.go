package dto

import "github.com/haschi/dinge/model"

type IndexFormData struct {
	Q      string
	S      string
	Result []model.DingRef
}
