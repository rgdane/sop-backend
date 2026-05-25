package dto

type DashboardCountsDto struct {
	SqlDivision  int64 `json:"sql_divisions"`
	SqlTitle     int64 `json:"sql_titles"`
	SqlFlowchart int64 `json:"sql_flowcharts"`
	SqlSop       int64 `json:"sql_sops"`
	SqlSopJob    int64 `json:"sql_sop_jobs"`
	SqlSpk       int64 `json:"sql_spks"`
	SqlSpkJob    int64 `json:"sql_spk_jobs"`
	SqlTotal     int64 `json:"sql_total"`

	GraphDivision  int64 `json:"graph_divisions"`
	GraphTitle     int64 `json:"graph_titles"`
	GraphFlowchart int64 `json:"graph_flowcharts"`
	GraphSop       int64 `json:"graph_sops"`
	GraphJob       int64 `json:"graph_jobs"`
	GraphSpk       int64 `json:"graph_spks"`
	GraphTotal     int64 `json:"graph_total"`
}
