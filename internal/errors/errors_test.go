package errors_test

import (
	"evm_event_indexer/internal/errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Errors(t *testing.T) {
	err := fmt.Errorf("test error")
	assert.Equal(t, "test error", err.Error())
	errW1 := errors.ErrInternalServerError.Wrap(err, "error wrap 1")
	errw2 := errors.ErrApiTimeout.Wrap(errW1)
	assert.Equal(t, "api timeout, something went wrong: error wrap 1, test error", errors.Chain(errw2))
}
