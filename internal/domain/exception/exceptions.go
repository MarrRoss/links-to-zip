package exception

import (
	"errors"
	"fmt"
)

var ErrGeneral = errors.New("general error")
var ErrDomain = fmt.Errorf("domain error: %w", ErrGeneral)
var ErrInvalidFileLink = fmt.Errorf("invalid file link: %w", ErrDomain)
var ErrInvalidArchivePath = fmt.Errorf("invalid archive path: %w", ErrDomain)
var ErrInvalidFileError = fmt.Errorf("invalid file error: %w", ErrDomain)
var ErrInvalidLinksCount = fmt.Errorf("invalid links count: %w", ErrDomain)
var ErrInvalidFileExtension = fmt.Errorf("invalid file extension: %w", ErrDomain)
var ErrInvalidURL = fmt.Errorf("invalid url format: %w", ErrDomain)
