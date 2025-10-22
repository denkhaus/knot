package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.UUID("project_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("title").
			MaxLen(200).
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Enum("state").
			Values("pending", "in-progress", "completed", "blocked", "cancelled", "deletion-pending").
			Default("pending"),
		field.Enum("priority").
			Values("low", "medium", "high").
			Default("medium"),
		field.Int("complexity").
			Min(1).
			Max(10),
		field.Int("depth").
			Default(0).
			NonNegative(),
		field.Int64("estimate").
			Optional().
			Nillable().
			Comment("Time estimate in minutes"),
		field.UUID("assigned_agent", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("completed_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		// Belongs to a project
		edge.From("project", Project.Type).
			Ref("tasks").
			Field("project_id").
			Unique().
			Required(),
		
		// Self-referencing parent-child relationship
		edge.To("children", Task.Type).
			From("parent").
			Field("parent_id").
			Unique(),
	}
}

// Indexes of the Task.
func (Task) Indexes() []ent.Index {
	return []ent.Index{
		// Core indexes for foreign keys and filtering
		index.Fields("project_id"),
		index.Fields("parent_id"),
		index.Fields("state"),
		index.Fields("priority"),
		index.Fields("assigned_agent"),
		index.Fields("complexity"),
		index.Fields("depth"),
		index.Fields("created_at"),
		
		// Composite indexes for common query patterns
		index.Fields("project_id", "state"),
		index.Fields("project_id", "priority"),
		index.Fields("project_id", "assigned_agent"),
		index.Fields("project_id", "parent_id"),
		index.Fields("project_id", "depth"),
		index.Fields("state", "complexity"),
		index.Fields("priority", "state"),
		index.Fields("priority", "complexity"),
	}
}
