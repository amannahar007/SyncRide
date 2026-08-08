package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/syncride/backend/auth-service/internal/domain"
	"github.com/syncride/backend/auth-service/internal/infrastructure/repository"
)

func TestTokenRepository_CreateRefreshToken(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewTokenRepository(db)
	ctx := context.Background()

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "hash123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "refresh_tokens" .+ RETURNING "id"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(token.ID))
		mock.ExpectCommit()

		err := repo.CreateRefreshToken(ctx, token)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTokenRepository_FindRefreshToken(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewTokenRepository(db)
	ctx := context.Background()

	hash := "hash123"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "token_hash"}).AddRow(uuid.New(), hash)

		mock.ExpectQuery(`SELECT \* FROM "refresh_tokens" WHERE token_hash = \$1 AND "refresh_tokens"\."deleted_at" IS NULL ORDER BY "refresh_tokens"\."id" LIMIT \$\d+`).
			WithArgs(hash, 1).
			WillReturnRows(rows)

		res, err := repo.FindRefreshToken(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, hash, res.TokenHash)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "refresh_tokens" WHERE token_hash = \$1 AND "refresh_tokens"\."deleted_at" IS NULL ORDER BY "refresh_tokens"\."id" LIMIT \$\d+`).
			WithArgs(hash, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindRefreshToken(ctx, hash)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTokenRepository_RevokeRefreshToken(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewTokenRepository(db)
	ctx := context.Background()

	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked"=$1,"updated_at"=$2 WHERE id = $3 AND "refresh_tokens"."deleted_at" IS NULL`)).
			WithArgs(true, sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RevokeRefreshToken(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTokenRepository_RevokeAllUserTokens(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewTokenRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked"=$1,"updated_at"=$2 WHERE (user_id = $3 AND revoked = $4) AND "refresh_tokens"."deleted_at" IS NULL`)).
			WithArgs(true, sqlmock.AnyArg(), userID, false).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RevokeAllUserTokens(ctx, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
