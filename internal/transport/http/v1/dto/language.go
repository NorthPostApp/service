package dto

import (
	"fmt"
	"strings"
)

type Language string

var SupportedLanguages = []Language{
	"ZH",
	"EN",
}

const (
	LanguageZH Language = "ZH"
	LanguageEN Language = "EN"
)

func (l Language) Validate() error {
	switch l {
	case LanguageZH, LanguageEN:
		return nil
	default:
		return fmt.Errorf("unsupported language: %s", l)
	}
}

func (l Language) Get() string {
	return strings.ToLower(string(l))
}
