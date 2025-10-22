package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TaskDependency holds the schema definition for the TaskDependency entity.
// This is a junction table for many-to-many task dependencies.
type TaskDependency struct {
	ent.Schema
}

// Fields of the TaskDependency.
func (TaskDependency) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.UUID("task_id", uuid.UUID{}),
		field.UUID("depends_on_task_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TaskDependency.
func (TaskDependency) Edges() []ent.Edge {
	return []ent.Edge{
		// The task that has the dependency
		edge.To("task", Task.Type).
			Field("task_id").
			Unique().
			Required(),
		
		// The task that is depended upon
		edge.To("depends_on_task", Task.Type).
			Field("depends_on_task_id").
			Unique().
			Required(),
	}
}

// Indexes of the TaskDependency.
func (TaskDependency) Indexes() []ent.Index {
	return []ent.Index{
		// Individual indexes for foreign keys
		index.Fields("task_id"),
		index.Fields("depends_on_task_id"),
		
		// Unique constraint to prevent duplicate dependencies
		index.Fields("task_id", "depends_on_task_id").Unique(),
		
		// Index for reverse lookups (what depends on this task)
		index.Fields("depends_on_task_id", "task_id"),
		
		// Index for time-based queries
		index.Fields("created_at"),
	}
}
