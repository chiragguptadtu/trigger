package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/store"
)

func TestCreateAndGetUser(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		email := uniqueEmail()
		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email:        email,
			Name:         "Alice",
			PasswordHash: "hashed",
			IsAdmin:      false,
		})
		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "Alice", user.Name)
		assert.False(t, user.IsAdmin)
		assert.True(t, user.IsActive)

		fetched, err := q.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)

		byID, err := q.GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, byID.Email)
	})
}

func TestUpdateUser(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email:        uniqueEmail(),
			Name:         "Bob",
			PasswordHash: "hashed",
			IsAdmin:      false,
		})
		require.NoError(t, err)

		updated, err := q.UpdateUser(ctx, store.UpdateUserParams{
			ID:       user.ID,
			Name:     "Bob Updated",
			IsAdmin:  true,
			IsActive: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "Bob Updated", updated.Name)
		assert.True(t, updated.IsAdmin)
	})
}

func TestDeactivateUser(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email:        uniqueEmail(),
			Name:         "Carol",
			PasswordHash: "hashed",
			IsAdmin:      false,
		})
		require.NoError(t, err)

		err = q.DeactivateUser(ctx, user.ID)
		require.NoError(t, err)

		fetched, err := q.GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.False(t, fetched.IsActive)
	})
}
