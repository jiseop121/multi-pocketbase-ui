package app

import "github.com/jiseop121/pbdash/internal/apperr"

func MapErrorToExitCode(err error) int {
	return apperr.ExitCode(err)
}
