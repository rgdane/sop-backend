package handlers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/service"
)

func GetGraphByIdHandler(elementId string, filter dto.GraphFilterDto) (interface{}, error) {
	{
		data, err := service.GetGraphById(elementId, filter)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func GetGraphByPropsHandler(payload dto.GraphFilterDto) (interface{}, error) {
	data, err := service.GetGraphByProps(payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetGraphByLabelHandler(label string, filter dto.GraphFilterDto) (interface{}, error) {
	data, error := service.GetGraphByLabel(label, filter)
	if error != nil {
		return nil, error
	}
	return data, nil
}

func GetSOPGraphHandler() (any, error) {
	data, err := service.GetSOPGraphs()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetDocumentGraphHandler(documentId string, filter dto.GraphFilterDto) (interface{}, error) {
	data, err := service.GetDocumentGraph(documentId, filter)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetCommentGraphHandler(commentId string) (any, error) {
	// Panggil service layer
	data, err := service.GetCommentGraph(commentId)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func CreateTableGraphHandler(parentId string, relation string, payload dto.ColumnDto) error {
	// TODO: add graph handler
	if err := service.CreateTableGraph(parentId, relation, payload); err != nil {
		return err
	}
	return nil
}

func CreateTextGraphHandler(elementId string, payload dto.TextDto) error {
	// TODO: add graph handler
	if err := service.CreateTextGraph(elementId, payload); err != nil {
		return err
	}
	return nil
}

func CreateCommentGraphHandler(elementId string, comment dto.CommentDto) error {
	if err := service.CreateCommentGraph(elementId, comment); err != nil {
		return err
	}
	return nil
}

func CreateGraphHandler(payload dto.GraphNode) (any, error) {
	record, err := service.CreateGraph(payload)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func BulkCreateGraphHandler(payload []dto.GraphNode) error {
	err := service.BulkCreateGraph(payload)
	if err != nil {
		return err
	}
	return nil
}

func CreateReviewGraphHandler(elementId string, review dto.ReviewDto) error {
	if err := service.CreateReviewGraph(elementId, review); err != nil {
		return err
	}
	return nil
}

func UpdateGraphHandler(elementId string, payload dto.NodeData) error {
	if err := service.UpdateGraph(elementId, payload); err != nil {
		return err
	}
	return nil
}

func UpdateMultipleGraphHandler(payload []dto.NodeData) error {
	if err := service.UpdateMultipleGraph(payload); err != nil {
		return err
	}
	return nil
}

func MergeGraphHandler(payload dto.GraphNode) error {
	if err := service.MergeGraph(payload); err != nil {
		return err
	}
	return nil
}

func UpdateTableGraphHandler(parentId string, relation string, payload dto.ColumnDto) error {
	if err := service.UpdateTableGraph(parentId, relation, payload); err != nil {
		return err
	}
	return nil
}

func DeleteGraphHandler(elementIds []string) error {
	if err := service.DeleteGraph(elementIds); err != nil {
		return err
	}
	return nil
}
