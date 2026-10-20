package cmd

import "github.com/xiaoxiaolin0918/nacos-cli/internal/risk"

// PrintedError means the JSON envelope was already written.
type PrintedError struct {
	Msg  string
	Code int
}

func (e *PrintedError) Error() string { return e.Msg }

func (e *PrintedError) ExitCode() int {
	if e.Code != 0 {
		return e.Code
	}
	return 1
}

var _ = risk.ExitConfirmationRequired
