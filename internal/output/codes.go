package output

// Stable machine-readable error codes for Agent / mgr consumers.
const (
	CodeConfirmationRequired = "confirmation_required"
	CodeServerUnset          = "server_unset"
	CodeCLIError             = "cli_error"
	CodeDoctorFailed         = "doctor_failed"
	CodeUpdateFailed         = "update_failed"
	CodeUnsupportedPlatform  = "unsupported_platform"
	CodePasswordRequired     = "password_required"
	CodeMissingFlag          = "missing_flag"
)

// HTTPCode returns "http_<status>" for a numeric HTTP status.
func HTTPCode(status int) string {
	if status < 100 || status > 599 {
		return "http_status"
	}
	return "http_" + itoa(status)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
