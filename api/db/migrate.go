package db

import (
	"github.com/chattycathy/api/db/models"
	"github.com/chattycathy/api/pkg/logger"
	"gorm.io/gorm"
)

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	logger.Info().Msg("Running database migrations...")

	err := db.AutoMigrate(
		&models.Ping{},
		&models.User{},
		&models.Permission{},
		&models.Role{},
		&models.UserRole{},
	)
	if err != nil {
		return err
	}

	// Seed default roles and permissions
	logger.Info().Msg("Seeding default roles and permissions...")
	if err := models.SeedDefaultRolesAndPermissions(db); err != nil {
		logger.Warn().Err(err).Msg("Failed to seed roles and permissions")
	}

	// Assign default role to existing users without any role
	logger.Info().Msg("Assigning default roles to users without roles...")
	if err := assignDefaultRoleToExistingUsers(db); err != nil {
		logger.Warn().Err(err).Msg("Failed to assign default roles to existing users")
	}

	logger.Info().Msg("Database migrations completed")
	return nil
}

// assignDefaultRoleToExistingUsers assigns the "user" role to users who don't have any role
func assignDefaultRoleToExistingUsers(db *gorm.DB) error {
	// Find all users without any role assigned
	var usersWithoutRoles []models.User
	err := db.Raw(`
		SELECT u.* FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.user_id IS NULL
	`).Scan(&usersWithoutRoles).Error
	if err != nil {
		return err
	}

	if len(usersWithoutRoles) == 0 {
		logger.Info().Msg("All users have roles assigned")
		return nil
	}

	logger.Info().Int("count", len(usersWithoutRoles)).Msg("Found users without roles")

	for _, user := range usersWithoutRoles {
		if err := models.AssignRoleToUser(db, user.ID, "user"); err != nil {
			logger.Warn().Err(err).Uint("user_id", user.ID).Msg("Failed to assign role to user")
		} else {
			logger.Info().Uint("user_id", user.ID).Str("email", user.Email).Msg("Assigned default 'user' role")
		}
	}

	return nil
}
