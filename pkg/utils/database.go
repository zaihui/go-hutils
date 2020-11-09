package utils

import (
	"time"

	"github.com/facebook/ent"
	"github.com/facebook/ent/dialect"
	"github.com/facebook/ent/schema/field"
	"github.com/facebook/ent/schema/mixin"
)

// BaseSchema 替换ent.Schema使用，包含套路字段
type BaseSchema struct {
	ent.Schema
}

func (BaseSchema) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

// BaseMixin  model补充一些套路字段
type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").MaxLen(32).Immutable(),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.MySQL: "datetime"}),
		field.Time("updated_at").Default(time.Now).SchemaType(map[string]string{dialect.MySQL: "datetime"}).UpdateDefault(time.Now),
		field.Time("deactivated_at").Nillable().SchemaType(map[string]string{dialect.MySQL: "datetime"}),
	}
}
