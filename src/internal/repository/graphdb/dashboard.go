package graphdb

import (
	"fmt"
	"jk-api/internal/shared/helper"
	"jk-api/pkg/neo4j/builder" // Sesuaikan dengan path package builder Anda
)

// GraphCountResult used for dashboard response mapping
type GraphCountResult struct {
	Division  int64 `json:"division"`
	Flowchart int64 `json:"flowchart"`
	Title     int64 `json:"title"`
	Sop       int64 `json:"sop"`
	Spk       int64 `json:"spk"`
	Job       int64 `json:"job"`
	TotalNode int64 `json:"total_node"`
}

type DashboardRepository interface {
	GetGraphCounts() (*GraphCountResult, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (repo *dashboardRepository) GetGraphCounts() (*GraphCountResult, error) {
	repoGraph := builder.NewGraphRepository()

	repoGraph = repoGraph.
		WithMatch("(n)").
		WithWhere("any(lbl IN labels(n) WHERE lbl IN ['Division', 'Flowchart', 'Title', 'SOP', 'SPK', 'Job'])", nil).
		WithReturn(`{
			division: count(CASE WHEN 'Division' IN labels(n) THEN 1 END),
			flowchart: count(CASE WHEN 'Flowchart' IN labels(n) THEN 1 END),
			title: count(CASE WHEN 'Title' IN labels(n) THEN 1 END),
			sop: count(CASE WHEN 'SOP' IN labels(n) THEN 1 END),
			spk: count(CASE WHEN 'SPK' IN labels(n) THEN 1 END),
			job: count(CASE WHEN 'Job' IN labels(n) THEN 1 END),
			total_node: count(n)
		} AS data`)

	records, err := repoGraph.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph counts: %w", err)
	}

if len(records) == 0 {
		return &GraphCountResult{}, nil
	}

	var result GraphCountResult

	if dataVal, ok := records[0].Get("data"); ok {
		if props, ok := dataVal.(map[string]any); ok {
			result.Division = helper.ToInt64(props["division"])
			result.Flowchart = helper.ToInt64(props["flowchart"])
			result.Title = helper.ToInt64(props["title"])
			result.Sop = helper.ToInt64(props["sop"])
			result.Spk = helper.ToInt64(props["spk"])
			result.Job = helper.ToInt64(props["job"])
			result.TotalNode = helper.ToInt64(props["total_node"])
		}
	}

	return &result, nil
}