package hutils

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
	datetime := map[string]string{dialect.MySQL: "datetime"}
	return []ent.Field{
		field.String("uid").DefaultFunc(NewUUID).MaxLen(32).MinLen(32).Unique().Immutable(),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(datetime),
		field.Time("updated_at").Default(time.Now).SchemaType(datetime).UpdateDefault(time.Now),
		field.Time("deactivated_at").Optional().Nillable().SchemaType(datetime),
	}
}
