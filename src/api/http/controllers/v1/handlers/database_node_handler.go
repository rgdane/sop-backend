package handlers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DatabaseNodeHandler struct {
	Service     service.DatabaseNodeService
}

func NewDatabaseNodeHandler(service service.DatabaseNodeService) *DatabaseNodeHandler {
	return &DatabaseNodeHandler{
		Service:     service,
	}
}

func (h *DatabaseNodeHandler) CreateDatabaseNodeHandler(input *dto.CreateDatabaseNodeDto, c *fiber.Ctx) (*dto.DatabaseNodeResponseDto, error) {
	var createdData *models.DatabaseNode

	err := h.Service.GetDB().Transaction(func(tx *gorm.DB) error {
		service := h.Service.WithTx(tx)

		payload, err := mapper.CreateDtoToDatabaseNode(input)
		if err != nil {
			return err
		}

		created, err := service.CreateDatabaseNode(payload)
		if err != nil {
			return err
		}
		createdData = created

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mapper.ToDatabaseNodeResponseDto(createdData)
}

func (h *DatabaseNodeHandler) UpdateDatabaseNodeHandler(id int64, input *dto.UpdateDatabaseNodeDto, c *fiber.Ctx) (*dto.DatabaseNodeResponseDto, error) {
	var updatedData *models.DatabaseNode

	err := h.Service.GetDB().Transaction(func(tx *gorm.DB) error {
		service := h.Service.WithTx(tx)

		payload, err := mapper.UpdateDtoToDatabaseNode(input)
		if err != nil {
			return err
		}

		updated, err := service.UpdateDatabaseNode(id, payload)
		if err != nil {
			return err
		}
		updatedData = updated

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mapper.ToDatabaseNodeResponseDto(updatedData)
}

func (h *DatabaseNodeHandler) DeleteDatabaseNodeHandler(id int64, c *fiber.Ctx) error {
	return h.Service.GetDB().Transaction(func(tx *gorm.DB) error {
		service := h.Service.WithTx(tx)

		if err := service.DeleteDatabaseNode(id); err != nil {
			return err
		}

		return nil
	})
}

func (h *DatabaseNodeHandler) GetDatabaseNodeByIDHandler(id int64, filter dto.DatabaseNodeFilter) (*dto.DatabaseNodeResponseDto, error) {
	data, err := h.Service.GetDatabaseNodeByID(id, filter)
	if err != nil {
		return nil, err
	}
	return mapper.ToDatabaseNodeResponseDto(data)
}

func (h *DatabaseNodeHandler) GetAllDatabaseNodesHandler(filter dto.DatabaseNodeFilter) ([]*dto.DatabaseNodeResponseDto, int64, error) {
	data, err := h.Service.GetAllDatabaseNodes(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := h.Service.CountDatabaseNodes(filter)
	if err != nil {
		return nil, 0, err
	}

	return mapper.ToDatabaseNodeResponseDtoList(data), total, nil
}
