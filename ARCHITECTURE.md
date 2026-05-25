# Architecture Document — SOP Backend

## Overview

A Go (Fiber) backend implementing **Clean Architecture** with **dual-database persistence** (PostgreSQL via GORM + Neo4j graph database). The project is organized into **API layer** and **Internal layer**, with dependency flow pointing inward.

```
┌─────────────────────────────────────────────┐
│                  API Layer                    │
│  DTO → Mapper → Controller → Handler         │
├─────────────────────────────────────────────┤
│               Internal Layer                  │
│  Container → Service → Repository (SQL/Graph)│
├─────────────────────────────────────────────┤
│            Database / Models                  │
│  GORM Models + Neo4j Nodes                   │
└─────────────────────────────────────────────┘
```

---

## Layer Breakdown

### 1. DTO (`src/api/http/controllers/v1/dto/`)
Request/response data transfer objects. Defines:
- `CreateSopJobDto` — input for creating a sop_job
- `UpdateSopJobDto` — partial update input (all pointer fields)
- `SopJobFilterDto` — comprehensive filter/pagination DTO
- `SopJobResponseDto` — response DTO embedding the model + optional relations
- Bulk variants: `BulkCreateSopJobs`, `BulkUpdateSopJobDto`, `BulkDeleteSopJobDto`
- `ReorderSopJobDto` — reordering payload

### 2. Mapper (`src/api/http/controllers/v1/mapper/`)
Transforms between DTO ↔ Model:
- `CreateSopJobDtoToModel` — maps create DTO → `models.SopJob`
- `UpdateSopJobDtoToModel` — maps update DTO → `map[string]any` (partial update map)
- `SopJobModelToResponseDto` — maps model → response DTO (with relations)

### 3. Controller (`src/api/http/controllers/v1/sop_job_controller.go`)
Fiber HTTP handlers. Every CRUD operation has **3 variants**:
| Variant | Route | Behavior |
|---------|-------|----------|
| **Hybrid** | `/sop-jobs` | SQL + Graph (transactional) |
| **SQL only** | `/sop-jobs/sql` | SQL only |
| **Graph only** | `/sop-jobs/graph` | Graph only |

Endpoints: `GET`, `GET/:id`, `POST`, `PUT/:id`, `DELETE/:id`, `POST /bulk-create`, `PUT /bulk-update`, `DELETE /bulk-delete`, `PUT /:id/reorder`.

Each controller:
1. Parses query params / body
2. Builds `SopJobFilterDto`
3. Delegates to `Handler`
4. Sends response via `presenters` package

### 4. Handler (`src/api/http/controllers/v1/handlers/sop_job_handler.go`)
Orchestrates the business flow:
- **Hybrid operations**: manages a **database transaction**, calls Service, then syncs to Neo4j, commits
- **Reference link generation**: when `Type + ReferenceID + Url` are present, prepends an HTML link to description
- **Bulk operations**: loops through DTOs, maps to models, calls service, syncs graph
- **Single-source handlers** (SQL/Graph only): simpler, skip the other database

### 5. Container (`src/internal/container/sop_job_container.go`)
**Dependency Injection** — wires up:
```
SQL Repository → Service → Handler
Graph Repository → Service → Handler
```
Returns `*handlers.SopJobHandler` ready to inject into controllers.

### 6. Service (`src/internal/service/sop_job_service.go`)
**Core business logic** implementing `SopJobService` interface:
- **SQL operations**: Create, Update (partial map), Delete (soft/hard), GetAll (with dynamic joins/filters), GetByID, Bulk ops, Reorder, Count
- **Graph operations**: delegates to `graphRepo`
- **Automatic Neo4j sync**: every SQL mutation triggers a corresponding graph mutation
- **Dynamic reference loading**: `loadDynamicReference` fetches `Sop` or `Spk` based on `Type`
- **Restore logic**: `sopJobRestore` re-inserts soft-deleted rows back to Neo4j with full relationship expansion via `apoc.path.expandConfig`
- **Exported helper**: `GetSOPJobGraphs()` — standalone function for Neo4j traversal queries

### 7. SQL Repository (`src/internal/repository/sql/sop_job.go`)
**GORM-based repository** with a **builder pattern**:
- Immutable `clone()` for each `With*` method
- `getQueryBuilder()` assembles select, joins, where, preloads, order, limit, cursor
- `FindSopJobWithJoins()` — custom join query returning `SopJobJoinResult` with title + reference (Sop/Spk) mappings
- `ReorderSopJob()` — manual index swap logic with shift-up/shift-down
- Uses a shared `pkg/gorm/builder` generic query builder

### 8. Graph Repository (`src/internal/repository/graphdb/sop_job.go`)
**Neo4j repository** using a Cypher query builder (`pkg/neo4j/builder`):
- `SopJobNode` — the graph node struct
- `GetAllGraphSopJobs` — MATCH + WHERE + Pattern Comprehension for title/reference names
- `Insert/Update/Bulk*` — MERGE/MATCH + SET operations
- `Delete` — soft-delete (`SET j.deleted_at`)
- `mapToSopJobNode` — manual property extraction from `map[string]any`

### 9. Model (`src/internal/database/models/sop_job.go`)
**GORM model** with:
- Composite indexes, sequences (`sop_jobs_seq`), auto-increment
- `BeforeCreate` hook: generates ID from sequence + code (`P%04d`)
- `AfterCreate` hook: auto-assigns `Index` (max+1), optionally links `Sop.parent_job_id`
- Relations: `HasSop`, `HasTitle`, `HasFlowchart`, `HasReference` (polymorphic interface{})

---

## Key Architectural Decisions

| Decision | Reasoning |
|----------|-----------|
| **Dual DB (SQL + Graph)** | SQL for relational queries, Neo4j for graph traversal (job → sop → title → etc.) |
| **Handler between Controller & Service** | Keeps controllers thin; handler manages transactions, graph sync, and cross-cutting concerns |
| **Partial update via `map[string]any`** | Enables precise partial updates without loading the full entity |
| **Builder pattern on repositories** | Immutable chaining for composable queries |
| **Soft delete by default** | `isPermanent` query param for hard delete |
| **Reference URL embedding** | Automatic link generation when creating/updating sop/SPK references |
| **Controller triplicates** | Flexibility to target a specific DB or both in a single request |

---

## Data Flow (Hybrid Create Example)

```
Client HTTP POST /sop-jobs
  → Controller (parse body → CreateSopJobDto)
    → Handler.CreateSopJobHandler()
      → Begin DB transaction
      → Mapper: DTO → Model
      → Service.CreateSopJob() → SQL Repo.InsertSopJob()
      → Auto-sync: Service.InsertGraphSopJob() → Graph Repo
      → Commit transaction
    → Mapper: Model → ResponseDto
  → Presenter: JSON response
```

## Filter DTO Fields

| Field | Type | Purpose |
|-------|------|---------|
| `Preload` | bool | Eager-load relations |
| `Type` | *string | Filter by job type |
| `SopID` | int64 | Filter by SOP |
| `SopName` | string | ILIKE search on SOP name |
| `TitleID` | int64 | Filter by title |
| `DivisionNames` | []string | Filter by division(s) |
| `Page/Limit` | int64 | Pagination |
| `Name` | string | ILIKE search on job name |
| `MinIndex` | int | `index > X` |
| `ReferenceID` | *int64 | Filter by reference |
| `ReferenceType` | string | sop or spk |
| `ShowDeleted` | bool | Include soft-deleted |
| `Sort/Order` | string | Sort field + direction |

## Endpoint Summary

| Method | Path | Scope |
|--------|------|-------|
| GET | `/sop-jobs` | Hybrid list |
| GET | `/sop-jobs/sql` | SQL only list |
| GET | `/sop-jobs/graph` | Graph only list |
| GET | `/sop-jobs/:id` | Hybrid by ID |
| GET | `/sop-jobs/sql/:id` | SQL by ID |
| GET | `/sop-jobs/graph/:id` | Graph by ID |
| POST | `/sop-jobs` | Hybrid create |
| POST | `/sop-jobs/sql` | SQL create |
| POST | `/sop-jobs/graph` | Graph create |
| PUT | `/sop-jobs/:id` | Hybrid update |
| PUT | `/sop-jobs/sql/:id` | SQL update |
| PUT | `/sop-jobs/graph/:id` | Graph update |
| DELETE | `/sop-jobs/:id` | Hybrid delete |
| DELETE | `/sop-jobs/sql/:id` | SQL delete |
| DELETE | `/sop-jobs/graph/:id` | Graph delete |
| POST | `/sop-jobs/bulk-create` | Hybrid bulk create |
| PUT | `/sop-jobs/bulk-update` | Hybrid bulk update |
| DELETE | `/sop-jobs/bulk-delete` | Hybrid bulk delete |
| PUT | `/sop-jobs/:id/reorder` | Reorder within SOP |

## Dependencies

- **Framework**: `github.com/gofiber/fiber/v2`
- **SQL ORM**: `gorm.io/gorm`
- **Graph DB**: Neo4j (via internal Cypher builder at `pkg/neo4j/builder`)
- **Architecture**: Clean Architecture with DI container
