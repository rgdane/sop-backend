package builder

import (
	"context"
	"fmt"
	"jk-api/internal/config"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// --- Interface ---
type GraphRepository interface {
	// --- CRUD operations ---
	FindNodes(label string) ([]neo4j.Record, error)
	FindNodeByID(label string, id string) (*neo4j.Record, error)
	CreateNode(label string, data map[string]interface{}) error
	UpdateNode(label string, id string, updates map[string]interface{}) error
	DeleteNode(label string, id string) error
	MergeNode(label string, id string, data map[string]interface{}) error

	// --- Fluent builder ---
	WithMatch(match string) GraphRepository
	WithMerge(merge string) GraphRepository
	WithCreate(create string) GraphRepository
	WithDelete(query string) GraphRepository
	WithDetachDelete(query string) GraphRepository
	WithWhere(where string, params map[string]interface{}) GraphRepository
	WithSet(set string, params map[string]interface{}) GraphRepository
	WithRemove(remove string, params map[string]interface{}) GraphRepository
	WithWith(query string) GraphRepository
	WithReturn(returns string) GraphRepository
	WithRelate(from, relation, to string, props map[string]interface{}) GraphRepository
	WithParams(params map[string]interface{}) GraphRepository
	WithLimit(limit int) GraphRepository
	WithOptionalMatch(query string) GraphRepository

	// --- APOC / Advanced Cypher ---
	WithCall(query string) GraphRepository
	WithYield(fields string) GraphRepository
	WithUnwind(variable string, asVar string) GraphRepository

	// --- Execution ---
	RunRead() ([]neo4j.Record, error)
	RunWrite() error
	RunWriteWithReturn() ([]neo4j.Record, error)
}

// --- Implementation ---
type graphRepository struct {
	driver        neo4j.DriverWithContext
	sessionConfig neo4j.SessionConfig
	ctx           context.Context

	statements []string
	params     map[string]interface{}
}

// --- Constructor ---
func NewGraphRepository() GraphRepository {
	return &graphRepository{
		driver:        config.GetNeo4j(),
		sessionConfig: neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite},
		ctx:           context.Background(),
		params:        make(map[string]interface{}),
		statements:    []string{},
	}
}

// --- Clone helper ---
func (r *graphRepository) clone() *graphRepository {
	clone := *r
	clone.params = make(map[string]interface{})
	for k, v := range r.params {
		clone.params[k] = v
	}
	clone.statements = append([]string{}, r.statements...)
	return &clone
}

// --- Builder methods ---
func (r *graphRepository) WithMatch(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "MATCH "+query)
	return clone
}

func (r *graphRepository) WithMerge(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "MERGE "+query)
	return clone
}

func (r *graphRepository) WithCreate(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "CREATE "+query)
	return clone
}

func (r *graphRepository) WithDelete(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "DELETE "+query)
	return clone
}

func (r *graphRepository) WithDetachDelete(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "DETACH DELETE "+query)
	return clone
}

func (r *graphRepository) WithWhere(query string, params map[string]interface{}) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "WHERE "+query)
	for k, v := range params {
		clone.params[k] = v
	}
	return clone
}

func (r *graphRepository) WithSet(query string, params map[string]interface{}) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "SET "+query)
	for k, v := range params {
		clone.params[k] = v
	}
	return clone
}

func (r *graphRepository) WithRemove(query string, params map[string]interface{}) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "REMOVE "+query)
	for k, v := range params {
		clone.params[k] = v
	}
	return clone
}

func (r *graphRepository) WithRelate(from, relation, to string, props map[string]interface{}) GraphRepository {
	clone := r.clone()

	stmt := fmt.Sprintf("MERGE (%s)-[:%s]->(%s)", from, relation, to)
	clone.statements = append(clone.statements, stmt)

	if len(props) > 0 {
		propStr := ""
		i := 0
		for k, v := range props {
			paramName := fmt.Sprintf("%s_param_%d", relation, i)
			if i > 0 {
				propStr += ", "
			}
			propStr += fmt.Sprintf("%s: $%s", k, paramName)
			clone.params[paramName] = v
			i++
		}
		stmt = fmt.Sprintf("MERGE (%s)-[:%s {%s}]->(%s)", from, relation, propStr, to)
		clone.statements[len(clone.statements)-1] = stmt
	}

	return clone
}

func (r *graphRepository) WithCall(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "CALL "+query)
	return clone
}

func (r *graphRepository) WithYield(fields string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "\nYIELD "+fields)
	return clone
}

func (r *graphRepository) WithUnwind(variable string, asVar string) GraphRepository {
	clone := r.clone()
	stmt := fmt.Sprintf("UNWIND %s AS %s", variable, asVar)
	clone.statements = append(clone.statements, stmt)
	return clone
}

func (r *graphRepository) WithReturn(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "RETURN "+query)
	return clone
}

func (r *graphRepository) WithParams(params map[string]interface{}) GraphRepository {
	clone := r.clone()
	for k, v := range params {
		clone.params[k] = v
	}
	return clone
}

func (r *graphRepository) WithLimit(limit int) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, fmt.Sprintf("LIMIT %d", limit))
	return clone
}

// --- Build ---
func (r *graphRepository) buildQuery() string {
	out := ""
	for _, s := range r.statements {
		out += s + "\n"
	}
	return out
}

func (r *graphRepository) WithOptionalMatch(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "OPTIONAL MATCH "+query)
	return clone
}

func (r *graphRepository) WithWith(query string) GraphRepository {
	clone := r.clone()
	clone.statements = append(clone.statements, "WITH "+query)
	return clone
}

// --- Executor ---
func (r *graphRepository) RunRead() ([]neo4j.Record, error) {
	query := r.buildQuery()

	log.Printf("[Neo4j READ] Executing query:\n%s\nParams: %#v\n", query, r.params)

	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	result, err := session.ExecuteRead(r.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(r.ctx, query, r.params)
		if err != nil {
			return nil, err
		}
		var records []neo4j.Record
		for res.Next(r.ctx) {
			records = append(records, *res.Record())
		}
		return records, res.Err()
	})
	if err != nil {
		return nil, err
	}
	return result.([]neo4j.Record), nil
}

func (r *graphRepository) RunWrite() error {
	query := r.buildQuery()

	log.Printf("[Neo4j WRITE] Executing query:\n%s\nParams: %#v\n", query, r.params)

	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	_, err := session.ExecuteWrite(r.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(r.ctx, query, r.params)
		return nil, err
	})
	return err
}

func (r *graphRepository) RunWriteWithReturn() ([]neo4j.Record, error) {
	query := r.buildQuery()

	log.Printf("[Neo4j WRITE WITH RETURN] Executing query:\n%s\nParams: %#v\n", query, r.params)

	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	result, err := session.ExecuteWrite(r.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(r.ctx, query, r.params)
		if err != nil {
			return nil, err
		}
		var records []neo4j.Record
		for res.Next(r.ctx) {
			records = append(records, *res.Record())
		}
		return records, res.Err()
	})
	if err != nil {
		return nil, err
	}
	return result.([]neo4j.Record), nil
}

// --- CRUD helper methods ---
func (r *graphRepository) FindNodes(label string) ([]neo4j.Record, error) {
	return r.
		WithMatch("(n:" + label + ")").
		WithReturn("n").
		RunRead()
}

func (r *graphRepository) FindNodeByID(label string, id string) (*neo4j.Record, error) {
	query := fmt.Sprintf("MATCH (n:%s {id: $id}) RETURN n", label)
	session := r.driver.NewSession(r.ctx, r.sessionConfig)
	defer session.Close(r.ctx)

	result, err := session.ExecuteRead(r.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(r.ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}
		if res.Next(r.ctx) {
			return res.Record(), nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	record := result.(*neo4j.Record)
	return record, nil
}

func (r *graphRepository) CreateNode(label string, data map[string]interface{}) error {
	return r.WithCreate("(n:"+label+")").
		WithSet("n = $data", map[string]interface{}{"data": data}).
		RunWrite()
}

func (r *graphRepository) MergeNode(label string, id string, data map[string]interface{}) error {
	return r.WithMerge("(n:"+label+" {id: $id})").
		WithSet("n += $data", map[string]interface{}{
			"id":   id,
			"data": data,
		}).
		RunWrite()
}

func (r *graphRepository) UpdateNode(label string, id string, updates map[string]interface{}) error {
	return r.WithMatch("(n:"+label+" {id: $id})").
		WithSet("n += $updates", map[string]interface{}{
			"id":      id,
			"updates": updates,
		}).
		RunWrite()
}

func (r *graphRepository) DeleteNode(label string, id string) error {
	return r.WithMatch("(n:" + label + " {id: $id})").
		WithReturn("n").
		RunWrite()
}
