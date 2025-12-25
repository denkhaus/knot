package mocks

//go:generate mockgen -source=../config/service.go -destination=mock_config_service.go -package=mocks Service
//go:generate mockgen -source=../logger/service.go -destination=mock_logger.go -package=mocks Logger
//go:generate mockgen -source=../manager/interfaces.go -destination=mock_manager.go -package=mocks ProjectManager
//go:generate mockgen -source=../types/types.go -destination=mock_repository_provider.go -package=mocks RepositoryProvider
//go:generate mockgen -source=../session/interface.go -destination=mock_session_manager.go -package=mocks Manager
//go:generate mockgen -source=../mcp/shared/registry.go -destination=mock_session_registry.go -package=mocks SessionRegistry
//go:generate mockgen -source=../mcp/transports/interface.go -destination=mock_transport.go -package=mocks Transport
//go:generate mockgen -source=../handlers/sync/interfaces.go -destination=mock_sync_handlers.go -package=mocks SyncService
//go:generate mockgen -source=../sync/shared/data_serializer.go -destination=mock_sync_data_serializer.go -package=mocks SyncDataSerializer
//go:generate mockgen -source=../sync/interfaces.go -destination=mock_sync_manager.go -package=mocks SyncManager

