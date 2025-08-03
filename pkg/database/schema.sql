-- RiskMatrix Database Schema

PRAGMA foreign_keys = ON;

-- Detections table
CREATE TABLE IF NOT EXISTS detections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    query TEXT, -- Detection query logic
    status TEXT NOT NULL, -- idea, draft, test, production, retired
    severity TEXT NOT NULL,
    risk_points INTEGER NOT NULL DEFAULT 0,
    playbook_link TEXT,
    owner TEXT,
    risk_object TEXT CHECK (risk_object IN ('IP', 'Host', 'User')),
    testing_description TEXT,
    event_count_last_30_days INTEGER NOT NULL DEFAULT 0,
    false_positives_last_30_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- MITRE ATT&CK Techniques
CREATE TABLE IF NOT EXISTS mitre_techniques (
    id TEXT PRIMARY KEY, -- e.g. T1059.001
    name TEXT NOT NULL,
    description TEXT,
    tactic TEXT NOT NULL, -- e.g. Execution
    tactics TEXT, -- JSON array of tactics
    domain TEXT, -- e.g. Enterprise, Mobile, ICS
    last_modified TEXT, -- Date when technique was last updated
    detection TEXT, -- Detection guidance
    platforms TEXT, -- JSON array of affected platforms
    data_sources TEXT, -- JSON array of data sources
    is_sub_technique BOOLEAN NOT NULL DEFAULT 0,
    sub_technique_of TEXT -- Parent technique ID if this is a sub-technique
);

-- Detection to MITRE Technique mapping
CREATE TABLE IF NOT EXISTS detection_mitre_map (
    detection_id INTEGER NOT NULL,
    mitre_id TEXT NOT NULL,
    PRIMARY KEY (detection_id, mitre_id),
    FOREIGN KEY (detection_id) REFERENCES detections(id) ON DELETE CASCADE,
    FOREIGN KEY (mitre_id) REFERENCES mitre_techniques(id) ON DELETE CASCADE
);

-- Data Sources
CREATE TABLE IF NOT EXISTS data_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    log_format TEXT
);

-- Detection to Data Source mapping
CREATE TABLE IF NOT EXISTS detection_datasource (
    detection_id INTEGER NOT NULL,
    datasource_id INTEGER NOT NULL,
    PRIMARY KEY (detection_id, datasource_id),
    FOREIGN KEY (detection_id) REFERENCES detections(id) ON DELETE CASCADE,
    FOREIGN KEY (datasource_id) REFERENCES data_sources(id) ON DELETE CASCADE
);

-- Risk Objects (entities that accumulate risk)
CREATE TABLE IF NOT EXISTS risk_objects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL, -- user, host, IP
    entity_value TEXT NOT NULL,
    current_score INTEGER NOT NULL DEFAULT 0,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_type, entity_value)
);

-- Events (detection triggers)
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    detection_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    raw_data TEXT, -- optional short blob or reference
    context TEXT, -- JSON field for detection context information
    risk_points INTEGER NOT NULL DEFAULT 0,
    is_false_positive BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (detection_id) REFERENCES detections(id) ON DELETE CASCADE,
    FOREIGN KEY (entity_id) REFERENCES risk_objects(id) ON DELETE CASCADE
);

-- Risk Alerts (generated when threshold exceeded)
CREATE TABLE IF NOT EXISTS risk_alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    triggered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total_score INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'New' CHECK (status IN ('New', 'Triage', 'Investigation', 'On Hold', 'Incident', 'Closed')),
    notes TEXT,
    owner TEXT,
    FOREIGN KEY (entity_id) REFERENCES risk_objects(id) ON DELETE CASCADE
);

-- False Positives
CREATE TABLE IF NOT EXISTS false_positives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    reason TEXT,
    analyst_name TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_detections_status ON detections(status);
CREATE INDEX IF NOT EXISTS idx_events_detection_id ON events(detection_id);
CREATE INDEX IF NOT EXISTS idx_events_entity_id ON events(entity_id);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_risk_objects_entity ON risk_objects(entity_type, entity_value);
CREATE INDEX IF NOT EXISTS idx_risk_alerts_entity_id ON risk_alerts(entity_id);
CREATE INDEX IF NOT EXISTS idx_false_positives_event_id ON false_positives(event_id);