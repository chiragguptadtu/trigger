package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/store"
)

func TestUpsertCommand(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		cmd, err := q.UpsertCommand(ctx, store.UpsertCommandParams{
			Slug:        "test-cmd",
			Name:        "Test Command",
			Description: "A test",
			ScriptPath:  "/commands/test_cmd.py",
		})
		require.NoError(t, err)
		assert.Equal(t, "test-cmd", cmd.Slug)
		assert.True(t, cmd.IsActive)

		// Upsert again — should update without error
		updated, err := q.UpsertCommand(ctx, store.UpsertCommandParams{
			Slug:        "test-cmd",
			Name:        "Test Command Renamed",
			Description: "Updated",
			ScriptPath:  "/commands/test_cmd.py",
		})
		require.NoError(t, err)
		assert.Equal(t, cmd.ID, updated.ID)
		assert.Equal(t, "Test Command Renamed", updated.Name)
	})
}

func TestListCommandsForUser_DirectGrant(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email: uniqueEmail(), Name: "Dave",
			PasswordHash: "hashed", IsAdmin: false,
		})
		require.NoError(t, err)

		cmd, err := q.UpsertCommand(ctx, store.UpsertCommandParams{
			Slug: "dave-cmd", Name: "Dave Cmd",
			Description: "", ScriptPath: "/commands/dave.py",
		})
		require.NoError(t, err)

		// No permission yet — list should be empty
		cmds, err := q.ListCommandsForUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, cmds)

		// Grant direct access
		_, err = q.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
			CommandID:   cmd.ID,
			GranteeType: "user",
			GranteeID:   user.ID,
		})
		require.NoError(t, err)

		cmds, err = q.ListCommandsForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, "dave-cmd", cmds[0].Slug)
	})
}

func TestListCommandsForUser_GroupGrant(t *testing.T) {
	withTx(t, func(q *store.Queries) {
		ctx := context.Background()

		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email: uniqueEmail(), Name: "Eve",
			PasswordHash: "hashed", IsAdmin: false,
		})
		require.NoError(t, err)

		group, err := q.CreateGroup(ctx, "ops-team")
		require.NoError(t, err)

		err = q.AddGroupMember(ctx, store.AddGroupMemberParams{
			UserID: user.ID, GroupID: group.ID,
		})
		require.NoError(t, err)

		cmd, err := q.UpsertCommand(ctx, store.UpsertCommandParams{
			Slug: "group-cmd", Name: "Group Cmd",
			Description: "", ScriptPath: "/commands/group.py",
		})
		require.NoError(t, err)

		// Grant access to the group
		_, err = q.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
			CommandID:   cmd.ID,
			GranteeType: "group",
			GranteeID:   group.ID,
		})
		require.NoError(t, err)

		cmds, err := q.ListCommandsForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, "group-cmd", cmds[0].Slug)
	})
}
