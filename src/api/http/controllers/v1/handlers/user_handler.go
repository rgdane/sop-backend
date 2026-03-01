package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"
)

type UserHandler struct {
	Service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) CreateUserHandler(input *dto.CreateUserDto) (*dto.UserResponseDto, error) {
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

	userService := h.Service.WithTx(db)

	payload, err := mapper.CreateUserDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := userService.CreateUser(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.UserModelToResponseDto(createdData)
}

func (h *UserHandler) UpdateUserHandler(id int64, input *dto.UpdateUserDto) (*models.User, error) {
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

	userService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateUserDtoToModel(input)
	updatedData, err := userService.UpdateUser(id, payload, associations)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *UserHandler) DeleteUserHandler(id int64, isPermanent bool) error {
	return h.Service.DeleteUser(id, isPermanent)
}

func (h *UserHandler) GetUserByIDHandler(id int64, filter dto.UserFilterDto) (*models.User, error) {
	return h.Service.GetUserByID(id, filter)
}

func (h *UserHandler) GetAllUsersHandler(filter dto.UserFilterDto) ([]models.User, int64, error) {
	data, err := h.Service.GetAllUsers(filter)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	db := h.Service.GetDB()
	if filter.Name != "" {
		db = db.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.ShowDeleted {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	}
	if err := db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *UserHandler) BulkCreateHandler(input *dto.BulkCreateUserDto) ([]*models.User, error) {
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

	userService := h.Service.WithTx(db)

	var users []*models.User
	for _, createDto := range input.Data {
		user, err := mapper.CreateUserDtoToModel(&createDto)
		if err != nil {
			return nil, err
		}
		if user != nil {
			users = append(users, user)
		}
	}

	if _, err := userService.BulkCreateUsers(users); err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return users, nil
}

func (h *UserHandler) BulkupdateHandler(input *dto.BulkUpdateUserDto) ([]*models.User, error) {
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

	userService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateUserDtoToModel(input.Data)
	if len(payload) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	// Use bulk update instead of per-ID updates to avoid FindByID with scoped queries
	err := userService.BulkUpdateUsers(input.IDs, payload, associations)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update users: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	// Return empty slice for now - bulk update succeeded
	return []*models.User{}, nil
}

func (h *UserHandler) BulkdeleteHandler(input *dto.BulkDeleteUserDto, isPermanent bool) error {
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

	userService := h.Service.WithTx(db)

	if err := userService.BulkDeleteUsers(input.IDs, isPermanent); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}
