package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/florin/utils"
	"github.com/noble-assets/florin/utils/mocks"
	"github.com/noble-assets/florin/x/florin/keeper"
	"github.com/noble-assets/florin/x/florin/types/blacklist"
	"github.com/stretchr/testify/require"
)

func TestBlacklistAcceptOwnership(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ACT: Attempt to accept ownership with no pending owner set.
	_, err := server.AcceptOwnership(goCtx, &blacklist.MsgAcceptOwnership{})
	// ASSERT: The action should've failed due to no pending owner set.
	require.ErrorIs(t, err, blacklist.ErrNoPendingOwner)

	// ARRANGE: Set pending owner in state.
	pendingOwner := utils.TestAccount()
	k.SetBlacklistPendingOwner(ctx, pendingOwner.Address)

	// ACT: Attempt to accept ownership with invalid signer.
	_, err = server.AcceptOwnership(goCtx, &blacklist.MsgAcceptOwnership{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidPendingOwner)

	// ACT: Attempt to accept ownership.
	_, err = server.AcceptOwnership(goCtx, &blacklist.MsgAcceptOwnership{
		Signer: pendingOwner.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.Equal(t, pendingOwner.Address, k.GetBlacklistOwner(ctx))
	require.Empty(t, k.GetBlacklistPendingOwner(ctx))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.OwnershipTransferred", events[0].Type)
}

func TestAddAdminAccount(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ACT: Attempt to add admin account with no owner set.
	_, err := server.AddAdminAccount(goCtx, &blacklist.MsgAddAdminAccount{})
	// ASSERT: The action should've failed due to no owner set.
	require.ErrorIs(t, err, blacklist.ErrNoOwner)

	// ARRANGE: Set owner in state.
	owner := utils.TestAccount()
	k.SetBlacklistOwner(ctx, owner.Address)

	// ACT: Attempt to add admin account with invalid signer.
	_, err = server.AddAdminAccount(goCtx, &blacklist.MsgAddAdminAccount{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidOwner)

	// ARRANGE: Generate an admin account.
	admin := utils.TestAccount()

	// ACT: Attempt to add admin account.
	_, err = server.AddAdminAccount(goCtx, &blacklist.MsgAddAdminAccount{
		Signer:  owner.Address,
		Account: admin.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.True(t, k.IsBlacklistAdmin(ctx, admin.Address))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.AdminAccountAdded", events[0].Type)
}

func TestBan(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ARRANGE: Set admin in state.
	admin := utils.TestAccount()
	k.SetBlacklistAdmin(ctx, admin.Address)

	// ACT: Attempt to ban with invalid signer.
	_, err := server.Ban(goCtx, &blacklist.MsgBan{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidAdmin)

	// ARRANGE: Generate an adversary account.
	adversary := utils.TestAccount()

	// ACT: Attempt to ban.
	_, err = server.Ban(goCtx, &blacklist.MsgBan{
		Signer:    admin.Address,
		Adversary: adversary.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.True(t, k.IsAdversary(ctx, adversary.Address))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.Ban", events[0].Type)
}

func TestRemoveAdminAccount(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ACT: Attempt to remove admin account with no owner set.
	_, err := server.RemoveAdminAccount(goCtx, &blacklist.MsgRemoveAdminAccount{})
	// ASSERT: The action should've failed due to no owner set.
	require.ErrorIs(t, err, blacklist.ErrNoOwner)

	// ARRANGE: Set owner in state.
	owner := utils.TestAccount()
	k.SetBlacklistOwner(ctx, owner.Address)

	// ACT: Attempt to remove admin account with invalid signer.
	_, err = server.RemoveAdminAccount(goCtx, &blacklist.MsgRemoveAdminAccount{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidOwner)

	// ARRANGE: Set admin in state.
	admin := utils.TestAccount()
	k.SetBlacklistAdmin(ctx, admin.Address)
	require.True(t, k.IsBlacklistAdmin(ctx, admin.Address))

	// ACT: Attempt to remove admin account.
	_, err = server.RemoveAdminAccount(goCtx, &blacklist.MsgRemoveAdminAccount{
		Signer:  owner.Address,
		Account: admin.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.False(t, k.IsBlacklistAdmin(ctx, admin.Address))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.AdminAccountRemoved", events[0].Type)
}

func TestBlacklistTransferOwnership(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ACT: Attempt to transfer ownership with no owner set.
	_, err := server.TransferOwnership(goCtx, &blacklist.MsgTransferOwnership{})
	// ASSERT: The action should've failed due to no owner set.
	require.ErrorIs(t, err, blacklist.ErrNoOwner)

	// ARRANGE: Set owner in state.
	owner := utils.TestAccount()
	k.SetBlacklistOwner(ctx, owner.Address)

	// ACT: Attempt to transfer ownership with invalid signer.
	_, err = server.TransferOwnership(goCtx, &blacklist.MsgTransferOwnership{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidOwner)

	// ACT: Attempt to transfer ownership to same owner.
	_, err = server.TransferOwnership(goCtx, &blacklist.MsgTransferOwnership{
		Signer:   owner.Address,
		NewOwner: owner.Address,
	})
	// ASSERT: The action should've failed due to same owner.
	require.ErrorIs(t, err, blacklist.ErrSameOwner)

	// ARRANGE: Generate a pending owner account.
	pendingOwner := utils.TestAccount()

	// ACT: Attempt to transfer ownership.
	_, err = server.TransferOwnership(goCtx, &blacklist.MsgTransferOwnership{
		Signer:   owner.Address,
		NewOwner: pendingOwner.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.Equal(t, owner.Address, k.GetBlacklistOwner(ctx))
	require.Equal(t, pendingOwner.Address, k.GetBlacklistPendingOwner(ctx))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.OwnershipTransferStarted", events[0].Type)
}

func TestUnban(t *testing.T) {
	k, ctx := mocks.FlorinKeeper()
	goCtx := sdk.WrapSDKContext(ctx)
	server := keeper.NewBlacklistMsgServer(k)

	// ARRANGE: Set admin in state.
	admin := utils.TestAccount()
	k.SetBlacklistAdmin(ctx, admin.Address)

	// ACT: Attempt to unban with invalid signer.
	_, err := server.Unban(goCtx, &blacklist.MsgUnban{
		Signer: utils.TestAccount().Address,
	})
	// ASSERT: The action should've failed due to invalid signer.
	require.ErrorIs(t, err, blacklist.ErrInvalidAdmin)

	// ARRANGE: Set adversary in state.
	adversary := utils.TestAccount()
	k.SetAdversary(ctx, adversary.Address)
	require.True(t, k.IsAdversary(ctx, adversary.Address))

	// ACT: Attempt to unban.
	_, err = server.Unban(goCtx, &blacklist.MsgUnban{
		Signer: admin.Address,
		Friend: adversary.Address,
	})
	// ASSERT: The action should've succeeded.
	require.NoError(t, err)
	require.False(t, k.IsAdversary(ctx, adversary.Address))
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "florin.blacklist.v1.Unban", events[0].Type)
}
