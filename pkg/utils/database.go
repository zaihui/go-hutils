package utils

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
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
	datatime := map[string]string{dialect.MySQL: "datetime"}
	return []ent.Field{
		field.String("uid").MaxLen(32).MinLen(32).Unique().Immutable(),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(datatime),
		field.Time("updated_at").Default(time.Now).SchemaType(datatime).UpdateDefault(time.Now),
		field.Time("deactivated_at").Optional().Nillable().SchemaType(datatime),
	}
}
