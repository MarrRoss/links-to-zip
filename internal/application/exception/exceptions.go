package exception

import (
	"fmt"
	"workmate_tz/internal/domain/exception"
)

var ErrApplication = fmt.Errorf("application error: %w", exception.ErrGeneral)
var MsgFileNotFound = fmt.Errorf("file not found in storage")
