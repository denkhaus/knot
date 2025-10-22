package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.String("title").
			MaxLen(200).
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Enum("state").
			Values("active", "completed", "archived", "deletion-pending").
			Default("active"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int("total_tasks").
			Default(0).
			NonNegative(),
		field.Int("completed_tasks").
			Default(0).
			NonNegative(),
		field.Float("progress").
			Default(0.0).
			Min(0.0).
			Max(100.0),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tasks", Task.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("title"),
		index.Fields("progress"),
		index.Fields("state"),
	}
}
