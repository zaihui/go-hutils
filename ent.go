package hutils

import (
	"errors"
	"time"
	"unicode/utf8"

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

// MaxRuneCount 字符串最大长度（包括中文）
func MaxRuneCount(maxLen int) func(s string) error {
	return func(s string) error {
		if utf8.RuneCountInString(s) > maxLen {
			return errors.New("value is more than the max length")
		}
		return nil
	}
}
