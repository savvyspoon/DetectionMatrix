package models

// MitreTechnique represents a MITRE ATT&CK technique
type MitreTechnique struct {
	ID             string   `json:"id"` // e.g. T1059.001
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Tactic         string   `json:"tactic"`                     // e.g. Execution
	Tactics        []string `json:"tactics,omitempty"`          // Multiple tactics a technique might belong to
	Domain         string   `json:"domain,omitempty"`           // e.g. Enterprise, Mobile, ICS
	LastModified   string   `json:"last_modified,omitempty"`    // Date when technique was last updated
	Detection      string   `json:"detection,omitempty"`        // Detection guidance
	Platforms      []string `json:"platforms,omitempty"`        // Affected platforms (Windows, macOS, Linux, etc.)
	DataSources    []string `json:"data_sources,omitempty"`     // Data sources useful for detection
	IsSubTechnique bool     `json:"is_sub_technique"`           // Whether this is a sub-technique
	SubTechniqueOf string   `json:"sub_technique_of,omitempty"` // Parent technique ID if this is a sub-technique
}

// MitreRepository defines the interface for MITRE technique data access
type MitreRepository interface {
	// Basic CRUD operations
	GetMitreTechnique(id string) (*MitreTechnique, error)
	ListMitreTechniques() ([]*MitreTechnique, error)
	ListMitreTechniquesByTactic(tactic string) ([]*MitreTechnique, error)
	CreateMitreTechnique(technique *MitreTechnique) error
	UpdateMitreTechnique(technique *MitreTechnique) error
	DeleteMitreTechnique(id string) error

	// Analytics
	GetCoverageByTactic() (map[string]float64, error)
	GetDetectionsByTechnique(techniqueID string) ([]*Detection, error)
}
