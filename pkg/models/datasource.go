package models

// DataSource represents a log source or telemetry feed
type DataSource struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`           // e.g. sysmon, cloudtrail
	Description string `json:"description,omitempty"`
	LogFormat   string `json:"log_format,omitempty"`
}

// DataSourceRepository defines the interface for data source data access
type DataSourceRepository interface {
	// Basic CRUD operations
	GetDataSource(id int64) (*DataSource, error)
	GetDataSourceByName(name string) (*DataSource, error)
	ListDataSources() ([]*DataSource, error)
	CreateDataSource(dataSource *DataSource) error
	UpdateDataSource(dataSource *DataSource) error
	DeleteDataSource(id int64) error
	
	// Analytics
	GetDetectionsByDataSource(dataSourceID int64) ([]*Detection, error)
	GetDataSourceUtilization() (map[string]int, error) // Returns count of detections per data source
}