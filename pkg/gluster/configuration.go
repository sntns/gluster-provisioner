package gluster

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Configuration struct {
	Host string `json:"host"`
}

func (c Configuration) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Host, validation.Required),
	)
}
