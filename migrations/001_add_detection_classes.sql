-- Migration: Add Detection Classes
-- Version: 001
-- Date: 2025-08-16
-- Description: Adds detection_classes table and class_id column to detections

-- Enable foreign keys
PRAGMA foreign_keys = ON;

BEGIN TRANSACTION;

-- Create detection_classes table if it doesn't exist
CREATE TABLE IF NOT EXISTS detection_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT, -- Hex color for UI display
    icon TEXT, -- Icon name for UI display
    is_system BOOLEAN NOT NULL DEFAULT 0, -- System defaults cannot be deleted
    display_order INTEGER NOT NULL DEFAULT 999,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Check if class_id column exists in detections table
-- SQLite doesn't support ALTER TABLE ADD COLUMN IF NOT EXISTS
-- So we need to check pragma first
-- This is handled programmatically in the application

-- Add class_id column to detections table (if not exists)
-- Note: SQLite doesn't support adding foreign key constraints to existing columns
-- The foreign key relationship will be enforced at the application level
ALTER TABLE detections ADD COLUMN class_id INTEGER REFERENCES detection_classes(id) ON DELETE SET NULL;

-- Insert default classes
INSERT OR IGNORE INTO detection_classes (name, description, color, icon, is_system, display_order) VALUES
    ('Auth', 'Authentication and authorization related detections', '#4CAF50', 'shield', 1, 1),
    ('Process', 'Process execution and manipulation detections', '#2196F3', 'cpu', 1, 2),
    ('Change', 'System and configuration change detections', '#FF9800', 'edit', 1, 3),
    ('Network', 'Network communication and traffic detections', '#9C27B0', 'network', 1, 4);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_detections_class_id ON detections(class_id);
CREATE INDEX IF NOT EXISTS idx_detection_classes_display_order ON detection_classes(display_order);

COMMIT;

-- Rollback script (save separately as 001_rollback.sql)
-- BEGIN TRANSACTION;
-- DROP INDEX IF EXISTS idx_detections_class_id;
-- DROP INDEX IF EXISTS idx_detection_classes_display_order;
-- -- Note: SQLite doesn't support DROP COLUMN, would need to recreate table
-- DROP TABLE IF EXISTS detection_classes;
-- COMMIT;