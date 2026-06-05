package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccessJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	SuccessJSON(c, "hello")

	assert.Equal(t, http.StatusOK, w.Code)
	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "hello", resp.Data)
}

func TestCreatedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	CreatedJSON(c, map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ErrorJSON(c, http.StatusBadRequest, "bad request")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "bad request", resp.Error)
}

func TestMessageJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	MessageJSON(c, http.StatusOK, "ok")

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Message)
}

func TestMessageJSON_Error(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	MessageJSON(c, http.StatusInternalServerError, "fail")

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "fail", resp.Message)
}

func TestPaginatedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	PaginatedJSON(c, []string{"a", "b"}, 1, 10, 25)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.Limit)
	assert.Equal(t, int64(25), resp.Meta.Total)
	assert.Equal(t, 3, resp.Meta.TotalPage)
}

func TestPaginatedJSON_ExactDivision(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	PaginatedJSON(c, []string{}, 1, 10, 20)

	var resp PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, resp.Meta.TotalPage)
}

func TestPaginatedJSON_ZeroTotal(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	PaginatedJSON(c, []string{}, 1, 10, 0)

	var resp PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp.Meta.TotalPage)
}
