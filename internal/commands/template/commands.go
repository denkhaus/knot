package template

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/templates"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// Commands returns all template-related CLI commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "list",
			Usage:  "List available task templates",
			Action: listAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "category",
					Aliases: []string{"c"},
					Usage:   "Filter by category",
				},
				&cli.StringSliceFlag{
					Name:    "tags",
					Aliases: []string{"t"},
					Usage:   "Filter by tags (can specify multiple)",
				},
				&cli.StringFlag{
					Name:    "search",
					Aliases: []string{"q"},
					Usage:   "Search in template names and descriptions",
				},
				&cli.BoolFlag{
					Name:  "built-in",
					Usage: "Show only built-in templates",
				},
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
			},
		},
		{
			Name:   "show",
			Usage:  "Show detailed information about a template",
			Action: showAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "Template name",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
			},
		},
		{
			Name:   "apply",
			Usage:  "Apply a template to create tasks",
			Action: applyAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "Template name",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "parent-id",
					Usage: "Parent task ID (optional)",
				},
				&cli.StringSliceFlag{
					Name:    "var",
					Aliases: []string{"v"},
					Usage:   "Template variables in format key=value (can specify multiple)",
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Preview what tasks would be created without actually creating them",
				},
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
			},
		},
		{
			Name:   "create",
			Usage:  "Create a new template from file",
			Action: createAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "file",
					Aliases:  []string{"f"},
					Usage:    "Template file path (YAML format)",
					Required: true,
				},
			},
		},
		{
			Name:   "validate",
			Usage:  "Validate a template file",
			Action: validateAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "file",
					Aliases:  []string{"f"},
					Usage:    "Template file path (YAML format)",
					Required: true,
				},
			},
		},
		{
			Name:   "info",
			Usage:  "Show detailed information about a template including source",
			Action: infoAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "Template name",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
			},
		},
		{
			Name:   "edit",
			Usage:  "Edit a user template in default editor",
			Action: editAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "Template name",
					Required: true,
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "Delete a user template",
			Action: deleteAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Usage:    "Template name",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "force",
					Usage: "Skip confirmation prompt",
				},
			},
		},
	}
}

// listAction lists available templates
func listAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Load all templates (user + built-in)
		templates, err := loadAllTemplates()
		if err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}

		// Apply filters
		filtered := applyTemplateFilters(templates, c)

		if len(filtered) == 0 {
			fmt.Println("No templates found matching the criteria.")
			return nil
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return outputTemplatesAsJSON(filtered)
		}

		fmt.Printf("Found %d template(s):\n\n", len(filtered))
		for _, template := range filtered {
			fmt.Printf("* %s\n", template.Name)
			fmt.Printf("  Category: %s\n", template.Category)
			if len(template.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(template.Tags, ", "))
			}
			fmt.Printf("  Description: %s\n", template.Description)
			fmt.Printf("  Tasks: %d\n", len(template.Tasks))
			if len(template.Variables) > 0 {
				fmt.Printf("  Variables: %d\n", len(template.Variables))
			}
			fmt.Println()
		}

		return nil
	}
}

// showAction shows detailed information about a template
func showAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		templateName := c.String("name")
		
		template, err := findTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return outputTemplateAsJSON(template)
		}

		fmt.Printf("Template: %s\n", template.Name)
		fmt.Printf("Category: %s\n", template.Category)
		if len(template.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(template.Tags, ", "))
		}
		fmt.Printf("Description: %s\n\n", template.Description)

		if len(template.Variables) > 0 {
			fmt.Println("Variables:")
			for _, variable := range template.Variables {
				fmt.Printf("  %s (%s)", variable.Name, variable.Type)
				if variable.Required {
					fmt.Print(" [required]")
				}
				if variable.DefaultValue != nil {
					fmt.Printf(" [default: %s]", *variable.DefaultValue)
				}
				fmt.Printf("\n    %s\n", variable.Description)
				if variable.Type == types.VarTypeChoice && len(variable.Options) > 0 {
					fmt.Printf("    Options: %s\n", strings.Join(variable.Options, ", "))
				}
			}
			fmt.Println()
		}

		fmt.Printf("Tasks (%d):\n", len(template.Tasks))
		for i, task := range template.Tasks {
			indent := ""
			if task.ParentID != nil {
				indent = "  "
			}
			fmt.Printf("%s%d. %s (ID: %s)\n", indent, i+1, task.Title, task.ID)
			fmt.Printf("%s   Complexity: %d", indent, task.Complexity)
			if task.Estimate != nil {
				fmt.Printf(" | Estimate: %d min", *task.Estimate)
			}
			fmt.Println()
			if task.Description != "" {
				fmt.Printf("%s   %s\n", indent, task.Description)
			}
			if len(task.Dependencies) > 0 {
				fmt.Printf("%s   Dependencies: %s\n", indent, strings.Join(task.Dependencies, ", "))
			}
			fmt.Println()
		}

		return nil
	}
}

// applyAction applies a template to create tasks
func applyAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectIDStr := c.String("project-id")
		if projectIDStr == "" {
			return fmt.Errorf("project-id is required")
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return fmt.Errorf("invalid project ID: %w", err)
		}

		templateName := c.String("name")
		template, err := findTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Parse variables
		variables := make(map[string]string)
		for _, varStr := range c.StringSlice("var") {
			parts := strings.SplitN(varStr, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format: %s (expected key=value)", varStr)
			}
			variables[parts[0]] = parts[1]
		}

		// Validate required variables
		if err := validateTemplateVariables(template, variables); err != nil {
			return fmt.Errorf("variable validation failed: %w", err)
		}

		// Parse parent ID if provided
		var parentID *uuid.UUID
		if parentIDStr := c.String("parent-id"); parentIDStr != "" {
			parsed, err := uuid.Parse(parentIDStr)
			if err != nil {
				return fmt.Errorf("invalid parent ID: %w", err)
			}
			parentID = &parsed
		}

		dryRun := c.Bool("dry-run")
		
		// Apply template
		result, err := applyTemplate(appCtx, template, projectID, parentID, variables, dryRun)
		if err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return outputApplyResultAsJSON(result)
		}

		if dryRun {
			fmt.Printf("Template '%s' would create %d tasks:\n\n", template.Name, len(result.CreatedTasks))
		} else {
			fmt.Printf("Template '%s' successfully created %d tasks:\n\n", template.Name, len(result.CreatedTasks))
		}

		for i, task := range result.CreatedTasks {
			indent := ""
			for j := 0; j < task.Depth; j++ {
				indent += "  "
			}
			fmt.Printf("%s%d. %s (ID: %s)\n", indent, i+1, task.Title, task.ID)
			fmt.Printf("%s   Complexity: %d | State: %s\n", indent, task.Complexity, task.State)
			if task.Description != "" {
				fmt.Printf("%s   %s\n", indent, task.Description)
			}
			fmt.Println()
		}

		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, errMsg := range result.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}

		return nil
	}
}

// createAction creates a new template from file
func createAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		filePath := c.String("file")
		
		template, err := loadTemplateFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Validate template
		if err := validateTemplate(template); err != nil {
			return fmt.Errorf("template validation failed: %w", err)
		}

		// TODO: Save to database when template repository is implemented
		fmt.Printf("Template '%s' loaded and validated successfully.\n", template.Name)
		fmt.Printf("Note: Template persistence not yet implemented.\n")

		return nil
	}
}

// validateAction validates a template file
func validateAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		filePath := c.String("file")
		
		template, err := loadTemplateFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		if err := validateTemplate(template); err != nil {
			return fmt.Errorf("template validation failed: %w", err)
		}

		fmt.Printf("Template '%s' is valid.\n", template.Name)
		fmt.Printf("  Tasks: %d\n", len(template.Tasks))
		fmt.Printf("  Variables: %d\n", len(template.Variables))

		return nil
	}
}
// infoAction shows detailed information about a template including source
func infoAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		templateName := c.String("name")
		
		template, err := findTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Determine source
		source := "built-in (embedded)"
		filePath := ""
		if !template.IsBuiltIn {
			source = "user"
			userPath, err := templates.GetUserTemplateFilePath(templateName)
			if err == nil {
				filePath = userPath
			}
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			info := map[string]interface{}{
				"template": template,
				"source":   source,
				"filePath": filePath,
			}
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal info to JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Template: %s\n", template.Name)
		fmt.Printf("Source: %s\n", source)
		if filePath != "" {
			fmt.Printf("File Path: %s\n", filePath)
		}
		fmt.Printf("Category: %s\n", template.Category)
		if len(template.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(template.Tags, ", "))
		}
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Printf("Tasks: %d\n", len(template.Tasks))
		fmt.Printf("Variables: %d\n", len(template.Variables))
		fmt.Printf("Built-in: %t\n", template.IsBuiltIn)

		return nil
	}
}

// editAction opens a user template in the default editor
func editAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		templateName := c.String("name")
		
		// Check if template exists
		template, err := findTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Only allow editing user templates
		if template.IsBuiltIn {
			return fmt.Errorf("cannot edit built-in template \"%s\". Use \"knot template create\" to create a user version first", templateName)
		}

		// Get file path
		filePath, err := templates.GetUserTemplateFilePath(templateName)
		if err != nil {
			return fmt.Errorf("failed to get template file path: %w", err)
		}

		// Get editor from environment or use default
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano" // Default editor
		}

		// Open editor
		cmd := exec.Command(editor, filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		fmt.Printf("Template \"%s\" edited successfully.\n", templateName)
		fmt.Printf("Use \"knot template validate --file %s\" to validate changes.\n", filePath)

		return nil
	}
}

// deleteAction deletes a user template
func deleteAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		templateName := c.String("name")
		force := c.Bool("force")
		
		// Check if template exists
		template, err := findTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Only allow deleting user templates
		if template.IsBuiltIn {
			return fmt.Errorf("cannot delete built-in template \"%s\"", templateName)
		}

		// Confirmation prompt (unless force flag is used)
		if !force {
			fmt.Printf("Are you sure you want to delete user template \"%s\"? (y/N): ", templateName)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" && response != "yes" && response != "YES" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Delete the template
		if err := templates.DeleteUserTemplate(templateName); err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}

		fmt.Printf("User template \"%s\" deleted successfully.\n", templateName)
		
		// Check if built-in version exists
		builtInTemplates, err := loadBuiltInTemplates()
		if err == nil {
			for _, builtIn := range builtInTemplates {
				if strings.EqualFold(builtIn.Name, templateName) {
					fmt.Printf("Note: Built-in template \"%s\" is still available.\n", templateName)
					break
				}
			}
		}

		return nil
	}
}
