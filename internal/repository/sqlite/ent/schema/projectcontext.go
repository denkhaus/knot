package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// ProjectContext holds the schema definition for the ProjectContext entity.
// This stores the currently selected project for the CLI context.
type ProjectContext struct {
	ent.Schema
}

// Fields of the ProjectContext.
func (ProjectContext) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable().
			Comment("Primary key - should always be 1 (singleton)"),
		field.UUID("selected_project_id", uuid.UUID{}).
			Comment("Currently selected project ID"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("When the selection was last updated"),
		field.String("updated_by").
			NotEmpty().
			Comment("Who updated the selection"),
	}
}

// Edges of the ProjectContext.
func (ProjectContext) Edges() []ent.Edge {
	return []ent.Edge{
		// Reference to the selected project
		edge.To("selected_project", Project.Type).
			Field("selected_project_id").
			Unique().
			Required(),
	}
}