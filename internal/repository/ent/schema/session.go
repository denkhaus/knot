package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.String("client_id").
			MaxLen(255).
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("last_activity").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("expires_at").
			Optional(),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.String("actor").
			MaxLen(255).
			Optional(),
		field.Enum("status").
			Values("active", "inactive", "expired").
			Default("active"),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("project", Project.Type).
			Unique(),
	}
}

// Indexes of the Session.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_id"),
		index.Fields("created_at"),
		index.Fields("last_activity"),
		index.Fields("expires_at"),
		index.Fields("status"),
		index.Fields("client_id", "status"),
	}
}