package apphostgae

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMustRegisterDelayedFunc(t *testing.T) {
	f := MustRegisterDelayedFunc("testFunc", func(_ context.Context) {})
	assert.NotNil(t, f)
}
