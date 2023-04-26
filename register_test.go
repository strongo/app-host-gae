package apphostgae

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMustRegisterDelayedFunc(t *testing.T) {
	f := MustRegisterDelayedFunc("testFunc", func() {})
	assert.NotNil(t, f)
}
