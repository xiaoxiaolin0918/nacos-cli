package cmd

import (
	"regexp"
	"strings"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
)

var httpStatusRE = regexp.MustCompile(`(?i)\bHTTP\s+(\d{3})\b`)

func classifyError(err error, level string) output.ErrorBody {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	body := output.ErrorBody{
		Type:    "cli",
		Code:    output.CodeCLIError,
		Message: msg,
	}
	if level != "" {
		body.Risk = level
	}
	if strings.Contains(msg, "NACOS_SERVER not set") {
		body.Code = output.CodeServerUnset
		return body
	}
	if m := httpStatusRE.FindStringSubmatch(msg); len(m) == 2 {
		body.Type = "http"
		body.Code = "http_" + m[1]
		return body
	}
	low := strings.ToLower(msg)
	if strings.Contains(low, "confirmation") || strings.Contains(low, "without --yes") {
		body.Type = "confirmation"
		body.Subtype = "confirmation_required"
		body.Code = output.CodeConfirmationRequired
	}
	return body
}
