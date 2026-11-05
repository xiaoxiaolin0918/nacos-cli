package cmd

import (
	"strings"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
)

func hydrateProfilePassword(dir, name string, p *config.Profile) {
	if p == nil || strings.TrimSpace(p.Password) != "" {
		return
	}
	c, ok, err := credstore.LoadLogin(credstore.ProfileLoginDir(dir, name))
	if err != nil || !ok {
		return
	}
	p.Password = c.Password
	if strings.TrimSpace(p.Username) == "" && c.Username != "" {
		p.Username = c.Username
	}
}

func persistProfilePassword(dir, name string, p *config.Profile, clear bool) error {
	ldir := credstore.ProfileLoginDir(dir, name)
	if clear {
		p.Password = ""
		return credstore.ClearLogin(ldir)
	}
	pass := p.Password
	if strings.TrimSpace(pass) == "" {
		// keep existing encrypted password; do not clear
		p.Password = ""
		return nil
	}
	user := strings.TrimSpace(p.Username)
	if user == "" {
		user = "nacos"
	}
	if err := credstore.SaveLogin(ldir, credstore.LoginPassword{
		Username:  user,
		Password:  pass,
		AutoLogin: true,
	}); err != nil {
		return err
	}
	p.Password = ""
	return nil
}

func profilePasswordSet(dir, name string, p config.Profile) bool {
	if strings.TrimSpace(p.Password) != "" {
		return true
	}
	return credstore.HasLogin(credstore.ProfileLoginDir(dir, name))
}

func redactedProfileView(dir, name string, p config.Profile) map[string]any {
	view := p.RedactedView()
	view["password_set"] = profilePasswordSet(dir, name, p)
	if view["password_set"] == true {
		view["password"] = "******"
	} else {
		view["password"] = ""
	}
	return view
}
