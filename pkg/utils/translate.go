package utils

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func Translate(messageID string, params map[string]interface{}, c *fiber.Ctx) string {
	var translate string
	var err error

	localizeConfig := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	if params != nil {
		localizeConfig.TemplateData = params
	}

	translate, err = fiberi18n.Localize(c, localizeConfig)
	if err != nil {
		return messageID
	}
	return translate
}

func TranslateSafe(messageID string, params map[string]interface{}, locale string) string {
	b := LoadBundle()

	localizer := i18n.NewLocalizer(b, locale)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: params,
	})
	if err != nil {
		return messageID
	}

	fmt.Println("MESSAGE : ", msg)
	return msg
}

var (
	bundle *i18n.Bundle
	once   sync.Once
)

// Exported function
func LoadBundle() *i18n.Bundle {
	once.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

		// Load your JSON translation files
		bundle.LoadMessageFile("./pkg/i18n/localize/en.json")
		bundle.LoadMessageFile("./pkg/i18n/localize/km.json")
		bundle.LoadMessageFile("./pkg/i18n/localize/zh.json")
	})
	return bundle
}
