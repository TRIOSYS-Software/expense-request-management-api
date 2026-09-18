package repositories

import (
	"errors"
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type RolesRepo struct {
	db *gorm.DB
}

func NewRolesRepo(db *gorm.DB) *RolesRepo {
	return &RolesRepo{db: db}
}

func (r *RolesRepo) userCountsByRole(roleIDs []uint) (map[uint]int64, error) {
	counts := make(map[uint]int64, len(roleIDs))
	if len(roleIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		RoleID uint
		Total  int64
	}
	if err := r.db.Model(&models.Users{}).
		Select("role_id, COUNT(*) AS total").
		Where("role_id IN ?", roleIDs).
		Group("role_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.RoleID] = row.Total
	}
	return counts, nil
}

func (r *RolesRepo) GetRoles() ([]models.Roles, error) {
	var roles []models.Roles
	if err := r.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}

	roleIDs := make([]uint, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	counts, err := r.userCountsByRole(roleIDs)
	if err != nil {
		return nil, err
	}

	// A role with no users is absent from the grouped result; the zero value is
	// already correct for it.
	for i := range roles {
		roles[i].UserCount = counts[roles[i].ID]
	}

	return roles, nil
}

func (r *RolesRepo) GetRoleByID(id uint) (*models.Roles, error) {
	var role models.Roles
	if err := r.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}

	counts, err := r.userCountsByRole([]uint{role.ID})
	if err != nil {
		return nil, err
	}
	role.UserCount = counts[role.ID]

	return &role, nil
}

func (r *RolesRepo) CreateRoleWithPermissions(role *models.Roles, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&models.Roles{}).Where("LOWER(name) = LOWER(?)", role.Name).Count(&exists).Error; err != nil {
			return err
		}
		if exists > 0 {
			return errors.New("role with the same name already exists")
		}

		if err := tx.Create(role).Error; err != nil {
			return err
		}

		if len(permissionIDs) > 0 {
			var permissions []models.Permissions
			if err := tx.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
				return err
			}
			if err := tx.Model(role).Association("Permissions").Replace(&permissions); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RolesRepo) UpdateRoleWithPermissions(role *models.Roles, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.Roles
		if err := tx.First(&existing, role.ID).Error; err != nil {
			return err
		}

		if existing.Name != role.Name {
			var cnt int64
			if err := tx.Model(&models.Roles{}).Where("LOWER(name) = LOWER(?) AND id <> ?", role.Name, role.ID).Count(&cnt).Error; err != nil {
				return err
			}
			if cnt > 0 {
				return errors.New("another role with the same name exists")
			}
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{"name": role.Name, "description": role.Description, "is_admin": role.IsAdmin}).Error; err != nil {
			return err
		}

		var permissions []models.Permissions
		if len(permissionIDs) > 0 {
			if err := tx.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&existing).Association("Permissions").Replace(&permissions); err != nil {
			return err
		}

		return nil
	})
}

func (r *RolesRepo) DeleteRoleIfUnused(roleID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Users{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("role is assigned to users and cannot be deleted")
		}

		if err := tx.Delete(&models.Roles{}, roleID).Error; err != nil {
			return err
		}
		return nil
	})
}
