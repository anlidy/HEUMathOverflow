package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJSON_SuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	JSON(c).Code(http.StatusOK).Message("success").Data(gin.H{"key": "val"}).Send()

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Errorf("expected code=200, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message='success', got %q", resp.Message)
	}
}

func TestJSON_WithPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	JSON(c).Code(200).Data([]int{1, 2}).Pagination(1, 10, 100).Send()

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Pagination == nil {
		t.Fatal("expected pagination to be set")
	}
	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 10 || resp.Pagination.Total != 100 {
		t.Errorf("unexpected pagination: %+v", resp.Pagination)
	}
}

func TestJSON_DefaultStatusCodes(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	JSON(c).Message("ok").Send()

	if w.Code != http.StatusOK {
		t.Errorf("expected default status 200, got %d", w.Code)
	}
	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != http.StatusOK {
		t.Errorf("expected default code=200, got %d", resp.Code)
	}
}
