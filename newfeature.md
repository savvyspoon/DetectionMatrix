# Detection Class Management Feature

## Overview
This feature introduces a configurable "class" field for detections, allowing users to categorize detections with predefined or custom-defined class values. The system will include default classes (Auth, Process, Change, Network) with the ability for users to add their own classifications through a settings interface.

## Business Requirements
- Add a `class` field to the Detection model and database schema
- Provide default classes: Auth, Process, Change, Network
- Allow users to define custom class values through a settings page
- Enable filtering and sorting by class in the detections list page
- Maintain data integrity with foreign key relationships

---

# Actions Feature Implementation Plan

## Overview
Add a new "Actions" entity to DetectionMatrix that defines what happens after detections trigger events. Actions will have a many-to-many relationship with Detections, allowing multiple detections to trigger the same action and detections to have multiple response actions.

## Feature Requirements
- **Action Entity Fields:**
  - `id`: Primary key (auto-increment)
  - `name`: Action name (required)
  - `description`: Description of what the action does
  - `url`: URL/link to external resources (optional)
  - `status`: Lifecycle status [ToDo, In Progress, Production, Retired]
  - `code`: Textarea field for action code/scripts
  - `owner`: Person responsible for the action
  - `created_at`: Timestamp
  - `updated_at`: Timestamp

- **Relationships:**
  - Many-to-many relationship between Actions and Detections
  - Join table: `detection_action_map`

## Implementation Tasks

### Phase 1: Database Changes

#### 1.1 Create Actions Table
```sql
-- Add to pkg/database/schema.sql
CREATE TABLE IF NOT EXISTS actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    url TEXT,
    status TEXT NOT NULL CHECK (status IN ('ToDo', 'In Progress', 'Production', 'Retired')),
    code TEXT,
    owner TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_actions_status ON actions(status);
CREATE INDEX IF NOT EXISTS idx_actions_owner ON actions(owner);
```

#### 1.2 Create Junction Table for Many-to-Many Relationship
```sql
-- Detection to Action mapping table
CREATE TABLE IF NOT EXISTS detection_action_map (
    detection_id INTEGER NOT NULL,
    action_id INTEGER NOT NULL,
    PRIMARY KEY (detection_id, action_id),
    FOREIGN KEY (detection_id) REFERENCES detections(id) ON DELETE CASCADE,
    FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE CASCADE
);

-- Create indexes for foreign keys
CREATE INDEX IF NOT EXISTS idx_detection_action_detection ON detection_action_map(detection_id);
CREATE INDEX IF NOT EXISTS idx_detection_action_action ON detection_action_map(action_id);
```

#### 1.3 Database Migration Script
Create migration script: `scripts/migrations/001_add_actions.sql`
- Check if tables exist before creating
- Handle rollback scenario
- Add sample data for testing

### Phase 2: Backend Implementation

#### 2.1 Create Action Model
**File:** `pkg/models/action.go`
```go
package models

import "time"

// ActionStatus represents the lifecycle stage of an action
type ActionStatus string

const (
    ActionStatusToDo       ActionStatus = "ToDo"
    ActionStatusInProgress ActionStatus = "In Progress"
    ActionStatusProduction ActionStatus = "Production"
    ActionStatusRetired    ActionStatus = "Retired"
)

// Action represents a response action for detections
type Action struct {
    ID          int64        `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    URL         string       `json:"url,omitempty"`
    Status      ActionStatus `json:"status"`
    Code        string       `json:"code,omitempty"`
    Owner       string       `json:"owner,omitempty"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
    
    // Relationships
    Detections  []Detection  `json:"detections,omitempty"`
}
```

#### 2.2 Create Action Repository
**File:** `internal/action/repository.go`
- Implement CRUD operations:
  - `GetAction(id int64) (*models.Action, error)`
  - `ListActions() ([]*models.Action, error)`
  - `ListActionsByStatus(status models.ActionStatus) ([]*models.Action, error)`
  - `CreateAction(action *models.Action) error`
  - `UpdateAction(action *models.Action) error`
  - `DeleteAction(id int64) error`
- Relationship operations:
  - `AddDetection(actionID, detectionID int64) error`
  - `RemoveDetection(actionID, detectionID int64) error`
  - `GetDetectionsByAction(actionID int64) ([]*models.Detection, error)`
  - `GetActionsByDetection(detectionID int64) ([]*models.Action, error)`

#### 2.3 Update Detection Model and Repository
**Update:** `pkg/models/detection.go`
- Add `Actions []Action` field to Detection struct

**Update:** `internal/detection/repository.go`
- Add methods to manage action relationships:
  - `AddAction(detectionID, actionID int64) error`
  - `RemoveAction(detectionID, actionID int64) error`
  - `loadActions(detection *models.Detection) error`
- Update `GetDetection()` to include actions

#### 2.4 Create Action API Handler
**File:** `pkg/api/action.go`
```go
package api

type ActionHandler struct {
    repo *action.Repository
}

// Implement standard REST endpoints:
// GET    /api/actions           - List all actions
// GET    /api/actions/{id}      - Get specific action
// POST   /api/actions           - Create new action
// PUT    /api/actions/{id}      - Update action
// DELETE /api/actions/{id}      - Delete action
// GET    /api/actions/{id}/detections - Get detections for action
// POST   /api/actions/{id}/detections/{detectionId} - Link detection
// DELETE /api/actions/{id}/detections/{detectionId} - Unlink detection
```

#### 2.5 Update Server Routes
**Update:** `pkg/api/server.go`
- Add action routes to router configuration
- Initialize ActionHandler with repository

### Phase 3: Frontend Implementation

#### 3.1 Create Actions List Page
**File:** `web/static/actions-list.html`
- Table view of all actions
- Columns: Name, Status, Owner, Description, # of Detections, Actions
- Filter by status
- Search functionality
- Add/Edit/Delete buttons
- Use HTMX for dynamic updates

#### 3.2 Create Action Detail Page
**File:** `web/static/actions-detail.html`
- Display all action fields
- Code editor for action code (using CodeMirror or similar)
- List of linked detections
- Add/Remove detection associations
- Edit mode toggle
- Use Alpine.js for state management

#### 3.3 Create Action Add/Edit Form
**File:** `web/static/actions-add.html`
- Form fields for all action properties
- Status dropdown
- Code textarea with syntax highlighting
- Multi-select for detections
- Form validation
- HTMX form submission

#### 3.4 Update Detection Pages
**Update:** `web/static/detections-detail.html`
- Add "Actions" section showing linked actions
- Add/Remove action functionality
- Display action status badges

**Update:** `web/static/detections-add.html`
- Add action selection during detection creation

#### 3.5 Update Navigation
**Update:** All HTML files with navigation
- Add "Actions" menu item between "Detections" and "MITRE ATT&CK"

#### 3.6 Create JavaScript Module
**File:** `web/static/js/actions.js`
- Action management functions
- Code editor initialization
- Form validation
- HTMX event handlers

#### 3.7 Update CSS
**Update:** `web/static/css/main.css`
- Add styles for action status badges
- Code editor styling
- Action cards/list styling

### Phase 4: Testing

#### 4.1 Unit Tests
**File:** `internal/action/repository_test.go`
- Test all CRUD operations
- Test relationship management
- Test error cases

**File:** `pkg/api/action_test.go`
- Test all API endpoints
- Test validation
- Test error responses

#### 4.2 Integration Tests
**File:** `scripts/test/test-actions.ps1`
- Test complete action workflow:
  1. Create action
  2. Link to detection
  3. Update action status
  4. Verify relationships
  5. Delete action

#### 4.3 UI Tests
- Manual testing checklist:
  - [ ] Create new action
  - [ ] Edit existing action
  - [ ] Link action to detection
  - [ ] Remove action from detection
  - [ ] Filter actions by status
  - [ ] Search actions
  - [ ] Delete action
  - [ ] Verify cascade deletes

### Phase 5: Documentation

#### 5.1 API Documentation
**Update:** `README.md`
Add Actions API section:
```markdown
### Actions
- `GET /api/actions` - List all actions
- `GET /api/actions/{id}` - Get a specific action
- `POST /api/actions` - Create a new action
- `PUT /api/actions/{id}` - Update an action
- `DELETE /api/actions/{id}` - Delete an action
- `GET /api/actions/{id}/detections` - Get linked detections
- `POST /api/actions/{id}/detections/{detectionId}` - Link detection to action
- `DELETE /api/actions/{id}/detections/{detectionId}` - Unlink detection from action
```

#### 5.2 Update CLAUDE.md
Add section about Actions:
- Explain purpose and relationship to detections
- Document status workflow
- Provide code examples

#### 5.3 Create User Guide
**File:** `docs/actions-guide.md`
- How to create actions
- Best practices for action code
- Status lifecycle explanation
- Examples of common actions

## Implementation Order

1. **Database Schema** (Phase 1)
   - Create tables and migrations
   - Test database changes

2. **Backend Models and Repository** (Phase 2.1-2.2)
   - Implement Action model
   - Create repository with CRUD operations

3. **API Layer** (Phase 2.4-2.5)
   - Implement REST endpoints
   - Add routes to server

4. **Basic UI** (Phase 3.1-3.3)
   - Create list and detail pages
   - Implement add/edit forms

5. **Integration** (Phase 2.3, 3.4-3.5)
   - Update Detection model and UI
   - Add navigation links

6. **Polish** (Phase 3.6-3.7)
   - Add JavaScript enhancements
   - Style improvements

7. **Testing** (Phase 4)
   - Write and run all tests
   - Fix any issues

8. **Documentation** (Phase 5)
   - Update all documentation
   - Create user guide

## Estimated Timeline

- **Week 1:** Database schema, backend models, repository
- **Week 2:** API implementation, basic UI pages
- **Week 3:** Integration with detections, UI polish
- **Week 4:** Testing, documentation, bug fixes

## Risk Considerations

1. **Database Migration:** Ensure backup before running migrations
2. **Performance:** Add appropriate indexes for queries
3. **Security:** Validate and sanitize action code input
4. **Compatibility:** Test with existing detection workflows
5. **UI Complexity:** Keep interface simple and intuitive

## Success Metrics

- All CRUD operations work for Actions
- Many-to-many relationship properly implemented
- UI provides smooth user experience
- No regression in existing functionality
- Comprehensive test coverage (>80%)
- Documentation complete and accurate

## Notes

- Consider adding action execution history tracking in future
- May want to add action templates/library
- Could integrate with automation platforms (SOAR)
- Consider adding action validation/testing capability

---

# Detection Class Management Feature - Full Implementation Plan

## Technical Architecture

### 1. Database Schema Changes

#### 1.1 New Tables

**detection_classes**
```sql
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
```

#### 1.2 Detection Table Modification
```sql
ALTER TABLE detections ADD COLUMN class_id INTEGER;
ALTER TABLE detections ADD FOREIGN KEY (class_id) REFERENCES detection_classes(id) ON DELETE SET NULL;
```

#### 1.3 Default Data Insertion
```sql
INSERT INTO detection_classes (name, description, color, icon, is_system, display_order) VALUES
    ('Auth', 'Authentication and authorization related detections', '#4CAF50', 'shield', 1, 1),
    ('Process', 'Process execution and manipulation detections', '#2196F3', 'cpu', 1, 2),
    ('Change', 'System and configuration change detections', '#FF9800', 'edit', 1, 3),
    ('Network', 'Network communication and traffic detections', '#9C27B0', 'network', 1, 4);
```

### 2. Backend Implementation

#### 2.1 Model Updates

**File:** `pkg/models/detection.go`
```go
// DetectionClass represents a category for detections
type DetectionClass struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    Description  string    `json:"description,omitempty"`
    Color        string    `json:"color,omitempty"`
    Icon         string    `json:"icon,omitempty"`
    IsSystem     bool      `json:"is_system"`
    DisplayOrder int       `json:"display_order"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Update Detection struct
type Detection struct {
    // ... existing fields ...
    ClassID *int64          `json:"class_id,omitempty"`
    Class   *DetectionClass `json:"class,omitempty"`
    // ... rest of fields ...
}
```

#### 2.2 Repository Interface Updates

**File:** `pkg/models/detection_repository.go`
```go
// Add to DetectionRepository interface
type DetectionRepository interface {
    // ... existing methods ...
    
    // Detection Class operations
    GetDetectionClass(id int64) (*DetectionClass, error)
    ListDetectionClasses() ([]*DetectionClass, error)
    CreateDetectionClass(class *DetectionClass) error
    UpdateDetectionClass(class *DetectionClass) error
    DeleteDetectionClass(id int64) error
    
    // Filtered list operations
    ListDetectionsByClass(classID int64) ([]*Detection, error)
}
```

#### 2.3 Repository Implementation

**File:** `internal/detection/repository.go`
Add implementation methods for class operations:
- CRUD operations for detection classes
- Join queries to include class data when fetching detections
- Validation to prevent deletion of system classes
- Update existing detection queries to include class information

#### 2.4 API Endpoints

**File:** `pkg/api/detection_class.go` (new file)
```go
package api

type DetectionClassHandler struct {
    repo models.DetectionRepository
}

// Implement REST endpoints:
// GET    /api/detection-classes           - List all classes
// POST   /api/detection-classes           - Create new class
// PUT    /api/detection-classes/{id}      - Update class
// DELETE /api/detection-classes/{id}      - Delete class
// GET    /api/detections?class_id={id}    - Filter detections by class
```

### 3. Frontend Implementation

#### 3.1 Settings Page for Class Management

**File:** `web/static/settings.html` (new file)
- Management interface for detection classes
- Table showing all classes with edit/delete options
- Form to add new classes
- Visual preview of class colors and icons
- Protection for system classes

#### 3.2 JavaScript Controller

**File:** `web/static/js/settings.js` (new file)
- CRUD operations for detection classes
- Form validation
- Visual feedback for operations
- Confirmation dialogs for deletions

#### 3.3 Detection List Updates

**File:** `web/static/detections-list.html`
Updates:
- Add class filter dropdown
- Display class badges with colors
- Enable sorting by class column
- Show class icon in listings

**File:** `web/static/js/detections-list.js`
Updates:
- Load detection classes for filter dropdown
- Implement class-based filtering
- Add sorting logic for class column
- Format class display with colors/icons

#### 3.4 Detection Add/Edit Forms

**Files:** `web/static/detections-add.html`, `web/static/detections-edit.html`
- Add class selection dropdown
- Display selected class with preview
- Validate class selection

### 4. CSS Styling

**File:** `web/static/css/main.css`
```css
/* Class badge styles */
.class-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 500;
}

/* Settings page styles */
.settings-section {
    background: #f8f9fa;
    border-radius: 8px;
    padding: 1.5rem;
}

.class-preview {
    display: inline-block;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 500;
}

/* Sortable headers */
th.sortable {
    cursor: pointer;
    user-select: none;
}

th.sortable:hover {
    background-color: #f5f5f5;
}
```

### 5. Migration Strategy

#### 5.1 Database Migration
**File:** `migrations/001_add_detection_classes.sql`
- Create detection_classes table
- Add class_id column to detections
- Insert default classes
- Create indexes for performance

#### 5.2 Data Migration
- All existing detections will have null class_id initially
- Admin can bulk-assign classes post-migration
- No data loss during migration

### 6. Testing Requirements

#### 6.1 Unit Tests
- Test CRUD operations for detection classes
- Test system class protection
- Test filtering and sorting logic
- Test cascade behavior

#### 6.2 Integration Tests
- Test API endpoints
- Test database constraints
- Test UI interactions
- Test migration script

#### 6.3 Test Scenarios
- Create custom class
- Edit custom class
- Attempt to delete system class (should fail)
- Delete custom class
- Filter detections by class
- Sort detections by class
- Assign class to detection
- Remove class from detection

### 7. Implementation Timeline

**Phase 1: Backend (2-3 days)**
- Database schema and migrations
- Model and repository implementation
- API endpoints

**Phase 2: Settings Page (2 days)**
- Settings UI for class management
- JavaScript functionality
- Validation and error handling

**Phase 3: Detection Integration (2 days)**
- Update detection list UI
- Add filtering and sorting
- Update add/edit forms

**Phase 4: Testing & Polish (2 days)**
- Comprehensive testing
- Bug fixes
- Documentation

**Total: 8-9 days**

### 8. Navigation Updates

Add "Settings" link to all navigation menus:
```html
<nav>
    <ul>
        <li><a href="index.html">Dashboard</a></li>
        <li><a href="detections-list.html">Detections</a></li>
        <li><a href="mitre.html">MITRE ATT&CK</a></li>
        <li><a href="datasources.html">Data Sources</a></li>
        <li><a href="events.html">Events</a></li>
        <li><a href="risk-objects.html">Risk Objects</a></li>
        <li><a href="alerts.html">Risk Alerts</a></li>
        <li><a href="settings.html">Settings</a></li> <!-- NEW -->
    </ul>
</nav>
```

### 9. API Documentation

**New Endpoints:**
```markdown
### Detection Classes
- `GET /api/detection-classes` - List all detection classes
- `POST /api/detection-classes` - Create new detection class
- `PUT /api/detection-classes/{id}` - Update detection class
- `DELETE /api/detection-classes/{id}` - Delete detection class

### Detection Filtering
- `GET /api/detections?class_id={id}` - Filter detections by class
- `GET /api/detections?sort=class` - Sort detections by class name
```

### 10. Success Metrics

- All detections can be classified
- Users can create custom classes within 2 minutes
- Filtering by class reduces detection list navigation time
- System classes remain protected
- Zero data loss during migration
- Sorting and filtering work seamlessly

### 11. Future Enhancements

1. **Class Analytics**
   - Dashboard widgets for detection distribution by class
   - Risk scoring weighted by class
   - Class-specific thresholds

2. **Advanced Features**
   - Class hierarchies
   - Class-based permissions
   - Import/export class definitions
   - Auto-classification suggestions

3. **Integration**
   - Class-based playbook routing
   - Class-specific alert templates
   - API integration for class management