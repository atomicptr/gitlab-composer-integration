package composer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJson(t *testing.T) {
	repository := Repository{}

	_, err := repository.ToJson()
	assert.Nil(t, err)
}
