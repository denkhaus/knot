package templates

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

// Service defines the template service interface.
// This abstracts template management to enable dependency injection and testing.
type Service interface {
	// Template loading methods
	LoadBuiltInTemplates() ([]*types.TaskTemplate, error)
	LoadUserTemplates() ([]*types.TaskTemplate, error)
	LoadTemplateFromFile(filePath string) (*types.TaskTemplate, error)
	GetTemplate(name string) (*types.TaskTemplate, error)

	// Template management methods
	CheckAndSeedIfNeeded() error
	SaveUserTemplate(template *types.TaskTemplate) error
	DeleteUserTemplate(templateName string) error
	UserTemplateExists(templateName string) bool
	GetUserTemplateFilePath(templateName string) (string, error)

	// Seeding and metadata methods
	LoadSeedingMetadata() (*SeedingMetadata, error)
	IsTemplateSeeded(templateID uuid.UUID) (bool, error)
	GetSeededTemplateInfo(templateID uuid.UUID) (*SeededTemplateInfo, error)
	UpdateSeededTemplate(template *types.TaskTemplate) error
	RemoveSeededTemplate(templateID uuid.UUID) error
	ListSeededTemplates() (map[string]SeededTemplateInfo, error)
	GetSeedingStats() (map[string]any, error)

	// Directory management
	GetUserTemplatesDir() (string, error)
	EnsureUserTemplatesDir() error
}

// serviceImpl is the private implementation of the Service interface
type serviceImpl struct {
	// No dependencies for now, but ready for future DI
}

// NewService creates a new template service instance.
// This follows the dependency injection pattern from di.md.
func NewService(injector do.Injector) (Service, error) {
	return &serviceImpl{}, nil
}

// ProvideService is a convenience wrapper for DI registration
func ProvideService(injector do.Injector) (Service, error) {
	return NewService(injector)
}

// Template loading methods

func (s *serviceImpl) LoadBuiltInTemplates() ([]*types.TaskTemplate, error) {
	return LoadBuiltInTemplates()
}

func (s *serviceImpl) LoadUserTemplates() ([]*types.TaskTemplate, error) {
	return LoadUserTemplates()
}

func (s *serviceImpl) LoadTemplateFromFile(filePath string) (*types.TaskTemplate, error) {
	return LoadTemplateFromFile(filePath)
}

func (s *serviceImpl) GetTemplate(name string) (*types.TaskTemplate, error) {
	// Try to load from user templates first
	userTemplates, err := s.LoadUserTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range userTemplates {
		if template.Name == name {
			return template, nil
		}
	}

	// If not found in user templates, try built-in templates
	builtInTemplates, err := s.LoadBuiltInTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range builtInTemplates {
		if template.Name == name {
			return template, nil
		}
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// Template management methods

func (s *serviceImpl) CheckAndSeedIfNeeded() error {
	return CheckAndSeedIfNeeded()
}

func (s *serviceImpl) SaveUserTemplate(template *types.TaskTemplate) error {
	return SaveUserTemplate(template)
}

func (s *serviceImpl) DeleteUserTemplate(templateName string) error {
	return DeleteUserTemplate(templateName)
}

func (s *serviceImpl) UserTemplateExists(templateName string) bool {
	return UserTemplateExists(templateName)
}

func (s *serviceImpl) GetUserTemplateFilePath(templateName string) (string, error) {
	return GetUserTemplateFilePath(templateName)
}

// Seeding and metadata methods

func (s *serviceImpl) LoadSeedingMetadata() (*SeedingMetadata, error) {
	return LoadSeedingMetadata()
}

func (s *serviceImpl) IsTemplateSeeded(templateID uuid.UUID) (bool, error) {
	return IsTemplateSeeded(templateID)
}

func (s *serviceImpl) GetSeededTemplateInfo(templateID uuid.UUID) (*SeededTemplateInfo, error) {
	return GetSeededTemplateInfo(templateID)
}

func (s *serviceImpl) UpdateSeededTemplate(template *types.TaskTemplate) error {
	return UpdateSeededTemplate(template)
}

func (s *serviceImpl) RemoveSeededTemplate(templateID uuid.UUID) error {
	return RemoveSeededTemplate(templateID)
}

func (s *serviceImpl) ListSeededTemplates() (map[string]SeededTemplateInfo, error) {
	return ListSeededTemplates()
}

func (s *serviceImpl) GetSeedingStats() (map[string]any, error) {
	return GetSeedingStats()
}

// Directory management

func (s *serviceImpl) GetUserTemplatesDir() (string, error) {
	return GetUserTemplatesDir()
}

func (s *serviceImpl) EnsureUserTemplatesDir() error {
	return EnsureUserTemplatesDir()
}