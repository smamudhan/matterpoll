package i18n

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

// PluginAPI is the plugin API interface required to TODO
type PluginAPI interface {
	GetUser(userID string) (*model.User, *model.AppError)
	LogWarn(msg string, keyValuePairs ...interface{})
	GetConfig() *model.Config
}

// Message is a string that can be localized.
//  https://pkg.go.dev/github.com/nicksnyder/go-i18n/v2/i18n?tab=doc#Message for more details.
type Message struct {
	*i18n.Message
}

// LocalizeConfig configures a call to the Localize method on Localizer.
// https://pkg.go.dev/github.com/nicksnyder/go-i18n/v2/i18n?tab=doc#LocalizeConfig
type LocalizeConfig struct {
	*i18n.LocalizeConfig
}

// Localizer provides Localize and MustLocalize methods that return localized messages.
// https://pkg.go.dev/github.com/nicksnyder/go-i18n/v2/i18n?tab=doc#Localizer
type Localizer struct {
	*i18n.Localizer
}

type Bundle struct {
	*i18n.Bundle
	api PluginAPI
}

// initBundle loads all localization files in i18n into a bundle and return this
func InitBundle(api PluginAPI, path string) (*Bundle, error) {
	bundle := &Bundle{
		Bundle: i18n.NewBundle(language.English),
		api:    api,
	}

	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open i18n directory")
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "active.") {
			continue
		}

		if file.Name() == "active.en.json" {
			continue
		}
		_, err = bundle.LoadMessageFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load message file %s", file.Name())
		}
	}

	return bundle, nil
}

// GetUserLocalizer returns a localizer that localizes in the users locale
func (b *Bundle) GetUserLocalizer(userID string) *Localizer {
	user, err := b.api.GetUser(userID)
	if err != nil {
		b.api.LogWarn("Failed get user's locale", "error", err.Error())
		return b.GetServerLocalizer()
	}

	l := i18n.NewLocalizer(b.Bundle, user.Locale)
	return &Localizer{Localizer: l}
}

// GetServerLocalizer returns a localizer that localizes in the server default client locale
func (b *Bundle) GetServerLocalizer() *Localizer {
	local := *b.api.GetConfig().LocalizationSettings.DefaultClientLocale

	l := i18n.NewLocalizer(b.Bundle, local)
	return &Localizer{Localizer: l}
}

// LocalizeDefaultMessage localizer the provided message
func (b *Bundle) LocalizeDefaultMessage(l *Localizer, m *Message) string {
	s, err := l.LocalizeMessage(m.Message)
	if err != nil {
		b.api.LogWarn("Failed to localize message", "message ID", m.ID, "error", err.Error())
		return ""
	}
	return s
}

// LocalizeWithConfig localizer the provided localize config
func (b *Bundle) LocalizeWithConfig(l *Localizer, lc *LocalizeConfig) string {
	s, err := l.Localize(lc.LocalizeConfig)
	if err != nil {
		b.api.LogWarn("Failed to localize with config", "error", err.Error())
		return ""
	}
	return s
}
