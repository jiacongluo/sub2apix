package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAccountDeleteErrorRouter() (*gin.Engine, *stubAdminService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/accounts/delete-error-status", handler.DeleteErrorStatusAccounts)
	return router, adminSvc
}

func TestDeleteErrorStatusAccountsDeletesAllErrorAccountsOnly(t *testing.T) {
	router, adminSvc := setupAccountDeleteErrorRouter()
	adminSvc.accounts = []service.Account{
		{ID: 101, Name: "broken-a", Status: service.StatusError},
		{ID: 102, Name: "healthy", Status: service.StatusActive},
		{ID: 103, Name: "broken-b", Status: service.StatusError},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/delete-error-status", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, service.StatusError, adminSvc.lastListAccounts.status)
	require.Equal(t, "", adminSvc.lastListAccounts.platform)
	require.Equal(t, "", adminSvc.lastListAccounts.accountType)
	require.Equal(t, "", adminSvc.lastListAccounts.search)
	require.Equal(t, int64(0), adminSvc.lastListAccounts.groupID)
	require.ElementsMatch(t, []int64{101, 103}, adminSvc.deletedAccountIDs)

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Total   int `json:"total"`
			Success int `json:"success"`
			Failed  int `json:"failed"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Equal(t, 2, payload.Data.Total)
	require.Equal(t, 2, payload.Data.Success)
	require.Equal(t, 0, payload.Data.Failed)
}
