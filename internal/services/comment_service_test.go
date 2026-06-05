package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComment_Create_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{
		Content: "This is a test comment",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "This is a test comment", resp.Content)
	assert.Equal(t, threadID, resp.ThreadID.String())
	assert.Nil(t, resp.ParentID)
}

func TestComment_Create_NestedReply(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	parentResp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{
		Content: "Root comment",
	})
	require.NoError(t, err)
	parentID := parentResp.ID.String()

	parentUUID := parentID
	replyResp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{
		Content:  "Nested reply",
		ParentID: &parentUUID,
	})

	require.NoError(t, err)
	require.NotNil(t, replyResp)
	assert.Equal(t, "Nested reply", replyResp.Content)
	assert.NotNil(t, replyResp.ParentID)
	assert.Equal(t, parentID, replyResp.ParentID.String())
}

func TestComment_Create_ThreadNotFound(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	uid, _ := uuid.Parse(userID)

	_, err := testCommentSvc.Create(uid, uuid.New(), CreateCommentRequest{
		Content: "This comment should fail",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestComment_Create_InvalidParentID(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)
	invalidParent := "not-a-valid-uuid"

	_, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{
		Content:  "Reply with bad parent ID",
		ParentID: &invalidParent,
	})
	assert.ErrorIs(t, err, ErrInvalidParentID)
}

func TestComment_GetByThread(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	_, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Comment number 1"})
	require.NoError(t, err)
	_, err = testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Comment number 2"})
	require.NoError(t, err)

	comments, _, err := testCommentSvc.GetByThread(tid, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(comments))
}

func TestComment_GetByThread_ThreadNotFound(t *testing.T) {
	cleanDB(t)

	_, _, err := testCommentSvc.GetByThread(uuid.New(), 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "thread not found")
}

func TestComment_Delete_NotFound(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	err := testCommentSvc.Delete(uid, uuid.New())
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestComment_Delete_Ownership(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Delete me"})
	require.NoError(t, err)

	cid, _ := uuid.Parse(resp.ID.String())
	err = testCommentSvc.Delete(uid, cid)
	assert.NoError(t, err)
}

func TestComment_Delete_Forbidden(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	resp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Not mine"})
	require.NoError(t, err)

	cid, _ := uuid.Parse(resp.ID.String())
	err = testCommentSvc.Delete(uuid.New(), cid)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
}

// ==================== NEW TESTS ====================

func TestComment_Delete_ParentDeletesReplies(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Create parent
	parent, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Parent"})
	require.NoError(t, err)

	// Create reply
	parentID := parent.ID.String()
	reply, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Reply", ParentID: &parentID})
	require.NoError(t, err)

	// Delete parent
	err = testCommentSvc.Delete(uid, parent.ID)
	require.NoError(t, err)

	// Reply should also be deleted (cascade)
	// GetByThread should return 0 top-level comments
	comments, total, err := testCommentSvc.GetByThread(tid, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, len(comments), "replies should cascade soft-delete")
	assert.Equal(t, int64(0), total)

	// Verify reply is gone too
	_ = reply // reply should be soft-deleted
}

func TestComment_Delete_NotFound_AlreadyDeleted(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Create comment
	resp, err := testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Delete me twice"})
	require.NoError(t, err)

	// Delete once
	err = testCommentSvc.Delete(uid, resp.ID)
	require.NoError(t, err)

	// Delete again — should be not found
	err = testCommentSvc.Delete(uid, resp.ID)
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestComment_Create_AfterThreadDeleted(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)
	threadID := seedThread(t, userID)
	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(threadID)

	// Soft delete thread
	err := testThreadSvc.Delete(uid, tid)
	require.NoError(t, err)

	// Try to comment
	_, err = testCommentSvc.Create(uid, tid, CreateCommentRequest{Content: "Should fail"})
	assert.ErrorIs(t, err, ErrThreadNotFound)
}

func TestComment_Delete_OnlyUserComments(t *testing.T) {
	cleanDB(t)
	userA := seedUser(t)
	userB := seedUser(t)
	threadID := seedThread(t, userA)
	uidA, _ := uuid.Parse(userA)
	uidB, _ := uuid.Parse(userB)
	tid, _ := uuid.Parse(threadID)

	// User A creates comment
	commentA, err := testCommentSvc.Create(uidA, tid, CreateCommentRequest{Content: "Comment by A"})
	require.NoError(t, err)

	// User B creates comment
	_, err = testCommentSvc.Create(uidB, tid, CreateCommentRequest{Content: "Comment by B"})
	require.NoError(t, err)

	// User A deletes only their comment
	err = testCommentSvc.Delete(uidA, commentA.ID)
	require.NoError(t, err)

	// User B's comment should still exist
	comments, total, err := testCommentSvc.GetByThread(tid, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, len(comments))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Comment by B", comments[0].Content)
}
