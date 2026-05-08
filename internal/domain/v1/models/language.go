package models

import (
	"fmt"
	"strings"
)

type Language string

const (
	LanguageZH Language = "zh"
	LanguageEN Language = "en"
)

func (l Language) Validate() error {
	switch l.Lower() {
	case LanguageZH, LanguageEN:
		return nil
	default:
		return fmt.Errorf("unsupported language: %s", l)
	}
}

func (l Language) Get() string {
	return strings.ToLower(string(l))
}

func (l Language) Lower() Language {
	return Language(l.Get())
}
