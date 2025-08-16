-- Rollback Migration: Remove Detection Classes
-- Version: 001
-- Date: 2025-08-16
-- Description: Rolls back detection_classes table and related changes

PRAGMA foreign_keys = ON;

BEGIN TRANSACTION;

-- Drop indexes
DROP INDEX IF EXISTS idx_detections_class_id;
DROP INDEX IF EXISTS idx_detection_classes_display_order;

-- SQLite doesn't support dropping columns directly
-- To remove class_id from detections, we would need to:
-- 1. Create a new table without class_id
-- 2. Copy data from old table
-- 3. Drop old table
-- 4. Rename new table

-- Create temporary table without class_id
CREATE TABLE detections_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    query TEXT,
    status TEXT NOT NULL,
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

-- Copy data from existing table (excluding class_id)
INSERT INTO detections_temp (
    id, name, description, query, status, severity, risk_points,
    playbook_link, owner, risk_object, testing_description,
    event_count_last_30_days, false_positives_last_30_days,
    created_at, updated_at
)
SELECT 
    id, name, description, query, status, severity, risk_points,
    playbook_link, owner, risk_object, testing_description,
    event_count_last_30_days, false_positives_last_30_days,
    created_at, updated_at
FROM detections;

-- Drop the old table
DROP TABLE detections;

-- Rename temp table to detections
ALTER TABLE detections_temp RENAME TO detections;

-- Recreate indexes that were on the original table
CREATE INDEX IF NOT EXISTS idx_detections_status ON detections(status);

-- Drop detection_classes table
DROP TABLE IF EXISTS detection_classes;

COMMIT;