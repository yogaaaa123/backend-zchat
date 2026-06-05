package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThread_Create_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	req := CreateThreadRequest{
		Title:   "My First Thread Title Here",
		Content: "This is the content of my first thread. It has enough characters to pass.",
	}

	resp, err := testThreadSvc.Create(uid, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.Title, resp.Title)
	assert.Equal(t, req.Content, resp.Content)
	assert.Equal(t, uid, resp.UserID)
}

func TestThread_GetByID_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)
	resp, err := testThreadSvc.GetByID(uid, tid)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, threadID, resp.ID.String())
}

func TestThread_GetByID_NotFound(t *testing.T) {
	cleanDB(t)
	_, err := testThreadSvc.GetByID(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestThread_GetAll_Pagination(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	uid, _ := uuid.Parse(userID)

	for i := 0; i < 3; i++ {
		_, err := testThreadSvc.Create(uid, CreateThreadRequest{
			Title:   "Test Thread " + uuid.New().String()[:8],
			Content: "Content for thread with enough characters for validation.",
		})
		require.NoError(t, err)
	}

	threads, total, err := testThreadSvc.GetAll(uuid.Nil, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(threads))
	assert.Equal(t, int64(3), total)
}

func TestThread_Update_Ownership(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testThreadSvc.Update(uid, tid, UpdateThreadRequest{
		Title:   "Updated Title Here",
		Content: "Updated content with enough characters for the test to pass.",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title Here", resp.Title)
}

func TestThread_Update_Forbidden(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	otherID := uuid.New()
	tid, _ := uuid.Parse(threadID)

	_, err := testThreadSvc.Update(otherID, tid, UpdateThreadRequest{
		Title:   "Hacked Title Here",
		Content: "Hacked content with enough characters for testing purposes.",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestThread_Delete_Ownership(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	err := testThreadSvc.Delete(uid, tid)
	assert.NoError(t, err)

	_, err = testThreadSvc.GetByID(uid, tid)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestThread_Search(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	uid, _ := uuid.Parse(userID)

	_, err := testThreadSvc.Create(uid, CreateThreadRequest{
		Title:   "Golang Programming Guide 2024",
		Content: "Learn how to write Go code effectively with best practices.",
	})
	require.NoError(t, err)

	results, total, err := testThreadSvc.Search(uuid.Nil, "Golang", 1, 10)
	require.NoError(t, err)
	assert.True(t, total > 0)
	assert.Contains(t, results[0].Title, "Golang")

	results, total, err = testThreadSvc.Search(uuid.Nil, "xyznonexistent12345xyz", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, 0, len(results))
}

func TestThread_Update_NotFound(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	_, err := testThreadSvc.Update(uid, uuid.New(), UpdateThreadRequest{
		Title:   "Gak ada",
		Content: "Thread ini gak pernah ada di database.",
	})
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestThread_Delete_NotFound(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	err := testThreadSvc.Delete(uid, uuid.New())
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestThread_Search_EmptyKeyword(t *testing.T) {
	cleanDB(t)

	results, total, err := testThreadSvc.Search(uuid.Nil, "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, 0, len(results))
}

// ==================== NEW THREAD TESTS ====================

func TestThread_Search_IsLiked_True(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Like thread
	_, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)

	// Search — is_liked should be true
	results, _, err := testThreadSvc.Search(uid, "Test Thread", 1, 10)
	require.NoError(t, err)
	assert.True(t, len(results) > 0)
	assert.True(t, results[0].IsLiked, "liked thread should show is_liked=true")
}

func TestThread_Search_IsLiked_False(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	_ = threadID

	// Search without liking
	results, _, err := testThreadSvc.Search(uid, "Test Thread", 1, 10)
	require.NoError(t, err)
	assert.True(t, len(results) > 0)
	assert.False(t, results[0].IsLiked, "unliked thread should show is_liked=false")
}

func TestThread_Search_IsLiked_DifferentUser(t *testing.T) {
	cleanDB(t)
	userA := seedUser(t)
	userB := seedUser(t)
	threadID := seedThread(t, userA)
	uidA, _ := uuid.Parse(userA)
	uidB, _ := uuid.Parse(userB)
	tid, _ := uuid.Parse(threadID)

	// User A likes
	_, err := testLikeSvc.Toggle(uidA, tid)
	require.NoError(t, err)

	// User B searches — thread should NOT show is_liked for user B
	results, _, err := testThreadSvc.Search(uidB, "Test Thread", 1, 10)
	require.NoError(t, err)
	assert.True(t, len(results) > 0)
	assert.False(t, results[0].IsLiked, "user B should see is_liked=false")

	// User A searches — thread SHOULD show is_liked for user A
	resultsA, _, err := testThreadSvc.Search(uidA, "Test Thread", 1, 10)
	require.NoError(t, err)
	assert.True(t, len(resultsA) > 0)
	assert.True(t, resultsA[0].IsLiked, "user A should see is_liked=true")
}

func TestThread_Delete_CascadeLikeCount(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Like
	_, err := testLikeSvc.Toggle(uid, tid)
	require.NoError(t, err)

	count, err := testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete thread
	err = testThreadSvc.Delete(uid, tid)
	require.NoError(t, err)

	// Like count should now be 0 (like was hard-deleted via Unscoped)
	count, err = testLikeRepo.CountByThread(tid)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "like count should be 0 after thread delete (Unscoped)")
}
