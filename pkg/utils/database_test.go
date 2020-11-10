package utils

import (
	"testing"

	"github.com/facebook/ent"
	"github.com/stretchr/testify/assert"
)

// User example schema
type User struct {
	BaseSchema
}

func (User) Fields() []ent.Field {
	return []ent.Field{}
}

func TestBaseSchema(t *testing.T) {
	assert.Equal(t, len(User{}.Mixin()[0].Fields()), 4)
}
