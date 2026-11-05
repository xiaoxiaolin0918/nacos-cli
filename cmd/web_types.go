package cmd

import (
	"strings"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

// profileWriteRequest is the JSON body for PUT /api/profiles/{name}
// and for draft POST /api/profiles/{name}/check-auth.
type profileWriteRequest struct {
	Server        string             `json:"server"`
	Host          string             `json:"host"`
	Port          int                `json:"port"`
	AuthType      string             `json:"authType"`
	Username      string             `json:"username"`
	Password      string             `json:"password"`
	PasswordSet   bool               `json:"password_set"` // true when client intends to set Password
	ClearPassword bool               `json:"clear_password"`
	Namespace     string             `json:"namespace"`
	ContextPath   string             `json:"contextPath"`
	Format        string             `json:"format"`
	Meta          config.ProfileMeta `json:"meta"`
}

func applyPasswordUpdate(existing config.Profile, req profileWriteRequest) config.Profile {
	if req.ClearPassword {
		existing.Password = ""
		return existing
	}
	if req.PasswordSet || req.Password != "" {
		existing.Password = req.Password
		return existing
	}
	return existing
}

// mergeProfileWrite overlays a write/draft request onto an existing profile (may be zero).
// Password: non-empty / password_set uses new password; clear_password clears;
// otherwise keeps the saved password (callers may then fall back to credstore token).
func mergeProfileWrite(existing config.Profile, req profileWriteRequest) (config.Profile, error) {
	p := existing
	if strings.TrimSpace(req.Server) != "" {
		p.Server = strings.TrimRight(strings.TrimSpace(req.Server), "/")
		p.Host = ""
		p.Port = 0
	}
	if strings.TrimSpace(req.Host) != "" {
		p.Host = strings.TrimSpace(req.Host)
		if strings.TrimSpace(req.Server) == "" {
			p.Server = ""
		}
	}
	if req.Port > 0 {
		p.Port = req.Port
	}
	if strings.TrimSpace(req.AuthType) != "" {
		p.AuthType = strings.TrimSpace(req.AuthType)
	}
	if p.AuthType == "" {
		p.AuthType = "nacos"
	}
	// Username is always taken from the form when a draft/write body is present.
	p.Username = strings.TrimSpace(req.Username)
	p.Namespace = config.NormalizeNamespaceID(req.Namespace)
	if strings.TrimSpace(req.ContextPath) != "" {
		if err := config.ValidateContextPath(req.ContextPath); err != nil {
			return p, err
		}
		p.ContextPath = strings.TrimSpace(req.ContextPath)
	}
	if strings.TrimSpace(req.Format) != "" {
		norm, err := config.NormalizeFormat(req.Format)
		if err != nil {
			return p, err
		}
		p.Format = norm
	}
	p = applyPasswordUpdate(p, req)
	return p, nil
}