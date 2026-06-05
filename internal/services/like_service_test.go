package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLike_Toggle_Like(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testLikeSvc.Toggle(uid, tid)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Liked)
	assert.Equal(t, 1, resp.LikeCount)
}

func TestLike_Toggle_Unlike(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.True(t, resp.Liked)

	resp, err = testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.Equal(t, 0, resp.LikeCount)
}

func TestLike_Toggle_ThreadNotFound(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	uid, _ := uuid.Parse(userID)

	_, err := testLikeSvc.Toggle(uid, uuid.New())
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestLike_Toggle_MultipleLikesSameUser(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp1, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.True(t, resp1.Liked)

	resp2, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.False(t, resp2.Liked)
	assert.Equal(t, 0, resp2.LikeCount)

	resp3, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.True(t, resp3.Liked)
	assert.Equal(t, 1, resp3.LikeCount)
}

func TestLike_Toggle_CountCorrect(t *testing.T) {
	cleanDB(t)
	userID1 := seedUser(t)
	userID2 := seedUser(t)
	threadID := seedThread(t, userID1)

	uid1, _ := uuid.Parse(userID1)
	uid2, _ := uuid.Parse(userID2)
	tid, _ := uuid.Parse(threadID)

	resp, err := testLikeSvc.Toggle(uid1, tid)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.LikeCount)

	resp, err = testLikeSvc.Toggle(uid2, tid)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.LikeCount)

	resp, err = testLikeSvc.Toggle(uid1, tid)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.LikeCount)
}

// ==================== NEW TESTS ====================

func TestLike_Duplicate_UniqueConstraint(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Create first like via service
	_, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)

	// Try to create duplicate like directly at repo level
	dup := &models.Like{
		ID:       uuid.New(),
		UserID:   uid,
		ThreadID: tid,
	}
	err = testLikeRepo.Create(dup)
	assert.Error(t, err, "unique constraint should prevent duplicate like")
}

func TestLike_Toggle_LikeAgain_AfterUnlike(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Like 1
	resp1, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.True(t, resp1.Liked)
	assert.Equal(t, 1, resp1.LikeCount)

	// Unlike 1
	resp2, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.False(t, resp2.Liked)
	assert.Equal(t, 0, resp2.LikeCount)

	// Like 2 (after unlike) — should work without unique constraint error
	resp3, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)
	assert.True(t, resp3.Liked)
	assert.Equal(t, 1, resp3.LikeCount)
}

func TestLike_Toggle_DeletedThread(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Soft delete thread
	err := testThreadSvc.Delete(uid, tid)
	require.NoError(t, err)

	// Try to like deleted thread
	_, err = testLikeSvc.Toggle(uid, tid)
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestLike_SameUser_DifferentThreads(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadA := seedThread(t, userID)
	threadB := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tidA, _ := uuid.Parse(threadA)
	tidB, _ := uuid.Parse(threadB)

	respA, err := testLikeSvc.Toggle(uid, tidA)
	require.NoError(t, err)
	assert.True(t, respA.Liked)
	assert.Equal(t, 1, respA.LikeCount)

	respB, err := testLikeSvc.Toggle(uid, tidB)
	require.NoError(t, err)
	assert.True(t, respB.Liked)
	assert.Equal(t, 1, respB.LikeCount)
}

func TestLike_Delete_NotFound(t *testing.T) {
	cleanDB(t)

	err := testLikeRepo.Delete(uuid.New())
	assert.NoError(t, err, "deleting non-existent like should not error")
}

func TestLike_Count_ZeroAfterAllDeleted(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Like
	_, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)

	// Verify count = 1
	count, err := testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Unlike
	_, err = testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)

	// Verify count = 0
	count, err = testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestLike_Count_MultipleUsers(t *testing.T) {
	cleanDB(t)
	userID1 := seedUser(t)
	userID2 := seedUser(t)
	userID3 := seedUser(t)
	threadID := seedThread(t, userID1)
	uid1, _ := uuid.Parse(userID1)
	uid2, _ := uuid.Parse(userID2)
	uid3, _ := uuid.Parse(userID3)
	tid, _ := uuid.Parse(threadID)

	// 3 users like
	testLikeSvc.Toggle(uid1, tid)
	testLikeSvc.Toggle(uid2, tid)
	testLikeSvc.Toggle(uid3, tid)

	count, err := testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 1 user unlikes
	testLikeSvc.Toggle(uid1, tid)

	count, err = testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
