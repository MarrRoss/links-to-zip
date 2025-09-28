package exception

import "workmate_tz/internal/domain/model"

type FileError struct {
	Link string
	Err  error
}

func ErrIDsToErrStructs(ids []model.ID, errMsg error) []FileError {
	errs := make([]FileError, 0, len(ids))
	for _, id := range ids {
		errs = append(errs, FileError{
			Link: id.String(),
			Err:  errMsg,
		})
	}
	return errs
}
