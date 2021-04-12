package helpers

import (
	"errors"
	"fmt"
)

const ()

func GenError(error_format string, error_txt ...interface{}) error {
	return errors.New(fmt.Sprintf(error_format, error_txt...))
}
