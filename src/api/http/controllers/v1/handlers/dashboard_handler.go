package handlers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/service"
)

type DashboardHandler struct {
	DashboardService service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{DashboardService: service}
}

func (h *DashboardHandler) GetDashboardCountsHandler() (dto.DashboardCountsDto, error) {
	sqlCounts, err := h.DashboardService.GetSqlCounts()
	if err != nil {
		return dto.DashboardCountsDto{}, err
	}

	graphCounts, err := h.DashboardService.GetGraphCounts()
	if err != nil {
		return dto.DashboardCountsDto{}, err
	}

	res := dto.DashboardCountsDto{}

	// Map SQL counts from first row
	if len(sqlCounts) > 0 {
		sc := sqlCounts[0]
		res.SqlDivision = sc.SqlDivision
		res.SqlTitle = sc.SqlTitle
		res.SqlFlowchart = sc.SqlFlowchart
		res.SqlSop = sc.SqlSop
		res.SqlSopJob = sc.SqlSopJob
		res.SqlSpk = sc.SqlSpk
		res.SqlSpkJob = sc.SqlSpkJob
		res.SqlTotal = sc.SqlDivision + sc.SqlTitle + sc.SqlFlowchart + sc.SqlSop + sc.SqlSopJob + sc.SqlSpk + sc.SqlSpkJob
	}

	// Map Graph counts directly
	if graphCounts != nil {
		res.GraphDivision = graphCounts.Division
		res.GraphFlowchart = graphCounts.Flowchart
		res.GraphTitle = graphCounts.Title
		res.GraphSop = graphCounts.Sop
		res.GraphSpk = graphCounts.Spk
		res.GraphJob = graphCounts.Job
		res.GraphTotal = graphCounts.TotalNode
	}

	return res, nil
}
