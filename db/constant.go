package db

const (
	connectionTimeout   = 90 // seconds
	dbReconnectionTime  = 2  // in seconds
	DatabaseName        = "hzp_task"
	migrationFolderPath = "file://server/repositories/db/migrations"
	EndpointPath        = "/api/v1/taskmanagement"
)

type RepositoryType string

const (
	RepositoryTypeDB       RepositoryType = "DB"
	RepositoryTypeFile     RepositoryType = "File"
	RepositoryTypeInMemory RepositoryType = "InMemory"
)
