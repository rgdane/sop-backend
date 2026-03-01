package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct {
	Service service.DepartmentService
}

func NewDepartmentHandler(service service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{Service: service}
}

func (h *DepartmentHandler) CreateDepartmentHandler(input *dto.CreateDepartmentDto) (*dto.DepartmentResponseDto, error) {
	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	newData, err := mapper.CreateDepartmentDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := departmentService.CreateDepartment(newData)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.DepartmentModelToResponseDto(createdData)
}

func (h *DepartmentHandler) UpdateDepartmentHandler(id int64, input *dto.UpdateDepartmentDto) (*models.Department, error) {
	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	payload, err := mapper.UpdateDepartmentDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := departmentService.UpdateDepartment(id, payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *DepartmentHandler) DeleteDepartmentHandler(id int64) error {
	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	if err := departmentService.DeleteDepartment(id); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *DepartmentHandler) GetDepartmentByIDHandler(id int64, filter dto.DepartmentFilterDto) (*models.Department, error) {
	return h.Service.GetDepartmentByID(id, filter)
}

func (h *DepartmentHandler) GetAllDepartmentsHandler(filter dto.DepartmentFilterDto) ([]models.Department, int64, error) {
	data, err := h.Service.GetAllDepartments(filter)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	db := h.Service.GetDB()
	if filter.Name != "" {
		db = db.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if err := db.Model(&models.Department{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return data, total, nil
}

func (h *DepartmentHandler) BulkCreateDepartmentsHandler(input *dto.BulkCreateDepartments, c *fiber.Ctx) ([]*models.Department, error) {
	if len(input.Data) == 0 {
		return nil, fmt.Errorf("bulk create data cannot be empty")
	}

	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	var departments []*models.Department
	for _, createDto := range input.Data {
		department, err := mapper.CreateDepartmentDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		departments = append(departments, department)
	}

	created, err := departmentService.BulkCreateDepartments(departments)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return created, nil
}

func (h *DepartmentHandler) BulkUpdateHandler(input *dto.BulkUpdateDepartmentDto, c *fiber.Ctx) ([]*models.Department, error) {
	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	updates, err := mapper.UpdateDepartmentDtoToModel(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to map update data: %w", err)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	err = departmentService.BulkUpdateDepartments(input.IDs, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update departments: %w", err)
	}

	updatedDepartments, err := departmentService.GetDepartmentsByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated departments: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedDepartments, nil
}

func (h *DepartmentHandler) BulkDeleteHandler(input *dto.BulkDeleteDepartmentDto, c *fiber.Ctx) error {
	db := h.Service.GetDB().Begin()
	committed := false

	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	departmentService := h.Service.WithTx(db)

	if err := departmentService.BulkDeleteDepartments(input.IDs); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}
