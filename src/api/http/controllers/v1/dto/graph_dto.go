package dto

type ColumnDto struct {
	JobId      string      `json:"jobId"`
	ElementId  string      `json:"elementId,omitempty"`
	Name       string      `json:"name,omitempty"`
	NodeType   string      `json:"nodeType,omitempty"`
	Options    any         `json:"options,omitempty"`
	TableRef   string      `json:"tableRef,omitempty"`
	GraphRef   string      `json:"graphRef,omitempty"`
	TitleId    int64       `json:"title_id,omitempty"`
	HasTable   []ColumnDto `json:"has_table,omitempty"`
	HasJob     []ColumnDto `json:"has_job,omitempty"`
	HasColumn  []ColumnDto `json:"has_column,omitempty"`
	Rows       []RowNode   `json:"rows,omitempty"`
	ProjectId  int64       `json:"projectId,omitempty"`
	ProductId  int64       `json:"productId,omitempty"`
	EpicId     int64       `json:"epicId,omitempty"`
	FeatureId  int64       `json:"featureId,omitempty"`
	DocumentId int64       `json:"documentId,omitempty"`
	Index      int         `json:"index,omitempty"`
}

type TextDto struct {
	Value      string `json:"value"`
	LinkGPT    string `json:"linkGPT,omitempty"`
	FeatureId  int64  `json:"featureId,omitempty"`
	EpicId     int64  `json:"epicId,omitempty"`
	ProductId  int64  `json:"productId,omitempty"`
	ProjectId  int64  `json:"projectId,omitempty"`
	DocumentId string `json:"documentId,omitempty"`
	FilterId   string `json:"filterId,omitempty"`
}

type CommentDto struct {
	Id   string `json:"id"`
	Text string `json:"text" validate:"required"`
}

type ReviewDto struct {
	Id       string `json:"reviewId"`
	Text     string `json:"value" validate:"required"`
	Isolated bool   `json:"isolated"`
}

type GraphNode struct {
	ElementId string   `json:"elementId"`
	Node      NodeData `json:"node"`
}

type NodeData struct {
	Id           any            `json:"id,omitempty"`
	ElementId    string         `json:"elementId,omitempty"`
	Labels       []string       `json:"labels,omitempty"`
	Props        map[string]any `json:"props,omitempty"`
	HasColumn    []ColumnDto    `json:"has_column,omitempty"`
	Relationship string         `json:"relationship,omitempty"`
}

type RowNode struct {
	ElementId  string `json:"elementId"`
	RelationId string `json:"relationId"`
	RowIndex   int    `json:"rowIndex"`
	GroupIndex int    `json:"groupIndex,omitempty"`
	NodeType   string `json:"nodeType"`
	Value      any    `json:"value"`
	Url        string `json:"url"`
	DocumentId string `json:"documentId"`
}

type GraphFilterDto struct {
	CommentId   string
	ProductId   int64
	ProjectId   int64
	EpicId      int64
	FeatureId   int64
	FilterId    string
	Type        string
	ProductName string
	DocumentId  string
	Name        string
	Multiple    bool
}

type DeleteDto struct {
	ElementIds []string `json:"elementIds" validate:"required,min=1"`
}
