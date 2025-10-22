# Postgres Repository Implementation - Projektplan

## Ziel
Implementation eines Postgres-basierten Repositories mit Ent Framework für das Project Management Tool.

## Detaillierte Checkliste

### Phase 1: Setup und Dependencies
- [ ] **1.1** Ent Framework zu go.mod hinzufügen
- [ ] **1.2** Postgres Driver (pgx) zu go.mod hinzufügen
- [ ] **1.3** Package-Struktur erstellen: `pkg/tools/project/repository/`
- [ ] **1.4** Ent Schema-Definitionen erstellen
- [ ] **1.5** Ent Code-Generation konfigurieren

### Phase 2: Schema Design
- [ ] **2.1** Project Entity Schema definieren
- [ ] **2.2** Task Entity Schema definieren
- [ ] **2.3** TaskDependency Junction-Table Schema definieren
- [ ] **2.4** Relationships und Constraints definieren
- [ ] **2.5** Indizes für Performance-kritische Felder definieren

### Phase 3: Migration Strategy (Option A - Ent Auto-Migration)
- [x] **3.1** Migration-Optionen evaluiert - Option A gewählt
- [ ] **3.2** Ent Auto-Migration konfigurieren
- [ ] **3.3** Schema-Synchronisation für Development-Environment
- [ ] **3.4** Auto-Migration-Tests erstellen

### Phase 4: Repository Implementation
- [ ] **4.1** Repository-Struktur und Options-Pattern implementieren
- [ ] **4.2** Connection-Management implementieren
- [ ] **4.3** Custom Error-Types definieren
- [ ] **4.4** Transaction-Support implementieren
- [ ] **4.5** Project CRUD-Operationen implementieren
- [ ] **4.6** Task CRUD-Operationen implementieren
- [ ] **4.7** Task-Hierarchy-Operationen implementieren
- [ ] **4.8** Task-Dependency-Management implementieren
- [ ] **4.9** Agent-Assignment-Operationen implementieren
- [ ] **4.10** Query-Operationen und Filter implementieren
- [ ] **4.11** Metrics und Progress-Calculation implementieren

### Phase 5: Testing und Integration
- [ ] **5.1** Unit-Tests für Repository-Operationen
- [ ] **5.2** Integration-Tests mit echter Postgres-DB
- [ ] **5.3** Performance-Tests für komplexe Queries
- [ ] **5.4** Transaction-Tests für komplexe Operationen
- [ ] **5.5** Error-Handling-Tests

### Phase 6: Documentation und Finalisierung
- [ ] **6.1** API-Dokumentation erstellen
- [ ] **6.2** Usage-Examples erstellen
- [ ] **6.3** Performance-Guidelines dokumentieren
- [ ] **6.4** Migration-Guide erstellen

## Package-Struktur
```
pkg/tools/project/repository/
├── PROJECT_PLAN.md           # Dieser Plan
├── options.go                # Options-Pattern für Repository-Config
├── errors.go                 # Custom Error-Types
├── repository.go             # Repository-Implementation
├── transactions.go           # Transaction-Management
├── migrations.go             # Ent Auto-Migration Setup
├── ent/                      # Ent-generierte Dateien
│   ├── schema/              # Schema-Definitionen
│   │   ├── project.go
│   │   ├── task.go
│   │   └── taskdependency.go
│   └── ...                  # Ent-generierte Files
└── testdata/                # Test-Fixtures
    └── migrations/          # Test-Migrations
```

## Migration Strategy - Option A (Ent Auto-Migration)

### Gewählte Strategie: Ent Auto-Migration

**Implementierungs-Details:**
- Ent Client wird mit AutoMigrate() konfiguriert
- Schema-Änderungen werden automatisch bei Startup angewendet
- Entwicklung wird durch automatische Synchronisation beschleunigt
- Keine manuellen Migration-Files erforderlich

**Konfiguration:**
```go
client, err := ent.Open("postgres", dsn)
if err != nil {
    return nil, err
}

// Auto-Migration aktivieren
if err := client.Schema.Create(ctx); err != nil {
    return nil, fmt.Errorf("failed creating schema resources: %v", err)
}
```

**Vorteile für unser Projekt:**
- Schnelle Entwicklungszyklen
- Automatische Schema-Synchronisation
- Weniger Boilerplate-Code
- Ent-native Lösung

**Considerations:**
- Für Production-Deployments sollte Auto-Migration optional sein
- Schema-Änderungen werden bei jedem Start überprüft
- Backup-Strategy für Production-Datenbanken empfohlen

## Schema-Design Details

### Project Table
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    total_tasks INTEGER DEFAULT 0,
    completed_tasks INTEGER DEFAULT 0,
    progress DECIMAL(5,2) DEFAULT 0.0
);
```

### Task Table
```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    state VARCHAR(20) NOT NULL DEFAULT 'pending',
    complexity INTEGER NOT NULL CHECK (complexity >= 1 AND complexity <= 10),
    depth INTEGER NOT NULL DEFAULT 0,
    estimate BIGINT, -- in minutes
    assigned_agent UUID,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP
);
```

### Task Dependencies Junction Table
```sql
CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(task_id, depends_on_task_id)
);
```

### Performance-Indizes
```sql
-- Task-Queries
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_state ON tasks(state);
CREATE INDEX idx_tasks_assigned_agent ON tasks(assigned_agent);
CREATE INDEX idx_tasks_complexity ON tasks(complexity);
CREATE INDEX idx_tasks_depth ON tasks(depth);

-- Dependencies
CREATE INDEX idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);

-- Composite Indizes für häufige Queries
CREATE INDEX idx_tasks_project_state ON tasks(project_id, state);
CREATE INDEX idx_tasks_project_agent ON tasks(project_id, assigned_agent);
```

## Transaction-Requirements

### Komplexe Operationen die Transactions benötigen:
1. **DeleteTaskSubtree** - Löschen von Task-Hierarchien
2. **BulkUpdateTasks** - Bulk-Updates von Tasks
3. **DuplicateTask** - Task-Duplikation mit Dependencies
4. **Project-Deletion** - Projekt mit allen Tasks löschen
5. **Task-Dependency-Management** - Zirkuläre Dependencies verhindern

## Error-Handling Strategy

### Custom Error-Types:
```go
type RepositoryError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int

const (
    ErrorTypeNotFound ErrorType = iota
    ErrorTypeConstraintViolation
    ErrorTypeCircularDependency
    ErrorTypeMaxDepthExceeded
    ErrorTypeMaxTasksExceeded
    ErrorTypeConnectionError
    ErrorTypeTransactionError
)
```

## Nächste Schritte

**Phase 1 ist bereit zum Start:**
1. Dependencies hinzufügen (ent, pgx)
2. Package-Struktur erstellen
3. Ent-Schema definieren
4. Auto-Migration konfigurieren
