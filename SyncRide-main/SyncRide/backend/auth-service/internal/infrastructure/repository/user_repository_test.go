package repository_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/syncride/backend/auth-service/internal/domain"
	"github.com/syncride/backend/auth-service/internal/infrastructure/repository"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return db, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users" .+ RETURNING "id"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(user.ID))
		mock.ExpectCommit()

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash"}).
			AddRow(user.ID, user.Email, user.PasswordHash)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$\d+`).
			WithArgs(user.Email, 1).
			WillReturnRows(rows)

		// Gorm Preload for Roles
		mock.ExpectQuery(`SELECT \* FROM "user_roles"`).WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}))

		res, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, res.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$\d+`).
			WithArgs(user.Email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByEmail(ctx, user.Email)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(id, "test@test.com")

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$\d+`).
			WithArgs(id, 1).
			WillReturnRows(rows)

		// Gorm Preload for Roles
		mock.ExpectQuery(`SELECT \* FROM "user_roles"`).WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}))

		res, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, res.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
