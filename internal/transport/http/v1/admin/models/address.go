package models

import (
	"fmt"
)

type Language string

var SupportedLanguages = []Language{
	"ZH",
	"EN",
}

type GetAddressesRequest struct {
	Language Language `json:"language" binding:"required"`
}

func (l Language) Validate() error {
	if !l.IsValid() {
		return fmt.Errorf("language [%s] is not supported", l)
	}
	return nil
}

func (l Language) IsValid() bool {
	for _, supportedLanguage := range SupportedLanguages {
		if l == supportedLanguage {
			return true
		}
	}
	return false
}
