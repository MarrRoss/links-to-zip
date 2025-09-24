package adapter

import (
	"fmt"
	"workmate_tz/internal/domain/exception"
)

var ErrStorage = fmt.Errorf("storage error: %w", exception.ErrGeneral)
var ErrTaskLimit = fmt.Errorf("active task limit exceeded: %w", ErrStorage)
