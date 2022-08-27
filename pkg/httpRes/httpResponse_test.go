package httpRes_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sonyamoonglade/notification-service/pkg/httpRes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInternal(t *testing.T) {

	w := httptest.NewRecorder()

	httpRes.Internal(w)

	expectedBody := "Internal error"
	expectedCode := 500

	actualBody, err := io.ReadAll(w.Body)

	assert.NoError(t, err)
	assert.Equal(t, expectedBody, string(actualBody))
	assert.Equal(t, expectedCode, w.Code)

}

func TestJson(t *testing.T) {

	prodLogger, err := zap.NewProduction()
	assert.NoError(t, err)
	logger := prodLogger.Sugar()

	w := httptest.NewRecorder()

	content := map[string]interface{}{
		"message": "some testing message",
	}
	expectedContent, _ := json.Marshal(content)

	httpRes.Json(logger, w, http.StatusOK, content)

	respBytes, err := io.ReadAll(w.Body)
	assert.NoError(t, err)

	assert.Equal(t, expectedContent, respBytes)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
}
