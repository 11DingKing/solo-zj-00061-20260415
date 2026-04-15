package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rameshsunkara/go-rest-api-example/internal/middleware"
	"github.com/rameshsunkara/go-rest-api-example/pkg/flightrecorder"
	"github.com/rameshsunkara/go-rest-api-example/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LogEntry struct {
	Level     string        `json:"level"`
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	Path      string        `json:"path"`
	UserAgent string        `json:"userAgent"`
	RespStatus int          `json:"respStatus"`
	ElapsedMs time.Duration `json:"elapsedMs"`
}

func TestRequestLogMiddleware(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		userAgent      string
		handler        func(c *gin.Context)
	}{
		{
			name:       "GET request with 200 OK",
			method:     http.MethodGet,
			path:       "/test/1",
			statusCode: http.StatusOK,
			userAgent:  "test-agent-1",
			handler: func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			},
		},
		{
			name:       "POST request with 201 Created",
			method:     http.MethodPost,
			path:       "/test/2",
			statusCode: http.StatusCreated,
			userAgent:  "test-agent-2",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"status": "created"})
			},
		},
		{
			name:       "GET request with 404 Not Found",
			method:     http.MethodGet,
			path:       "/test/notfound",
			statusCode: http.StatusNotFound,
			userAgent:  "test-agent-3",
			handler: func(c *gin.Context) {
				c.String(http.StatusNotFound, "Not Found")
			},
		},
		{
			name:       "DELETE request with 204 No Content",
			method:     http.MethodDelete,
			path:       "/test/delete",
			statusCode: http.StatusNoContent,
			userAgent:  "test-agent-4",
			handler: func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			},
		},
		{
			name:       "GET request with 500 Internal Server Error",
			method:     http.MethodGet,
			path:       "/test/error",
			statusCode: http.StatusInternalServerError,
			userAgent:  "test-agent-5",
			handler: func(c *gin.Context) {
				c.String(http.StatusInternalServerError, "Internal Error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var logBuf bytes.Buffer
			gin.SetMode(gin.TestMode)
			router := gin.New()

			lgr := logger.New("info", &logBuf)
			router.Use(middleware.RequestLogMiddleware(lgr, nil))

			router.Handle(tc.method, tc.path, tc.handler)

			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)
			req.Header.Set("User-Agent", tc.userAgent)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, tc.statusCode, resp.Code)

			logOutput := logBuf.String()
			require.NotEmpty(t, logOutput, "log should not be empty")

			var logEntry LogEntry
			err = json.Unmarshal([]byte(logOutput), &logEntry)
			require.NoError(t, err, "log should be valid JSON")

			assert.Equal(t, "info", logEntry.Level)
			assert.Equal(t, tc.method, logEntry.Method)
			assert.Equal(t, tc.path, logEntry.URL)
			assert.Equal(t, tc.path, logEntry.Path)
			assert.Equal(t, tc.userAgent, logEntry.UserAgent)
			assert.Equal(t, tc.statusCode, logEntry.RespStatus)
			assert.GreaterOrEqual(t, logEntry.ElapsedMs, time.Duration(0), "elapsed time should be non-negative")
		})
	}
}

func TestRequestLogMiddlewareWithSlowRequest(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp("", "test-traces-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	var logBuf bytes.Buffer
	lgr := logger.New("info", &logBuf)
	fr := flightrecorder.New(lgr, tempDir, time.Second, 1<<20)
	require.NotNil(t, fr, "flight recorder should be created successfully")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestLogMiddleware(lgr, fr))

	router.GET("/slow", func(ctx *gin.Context) {
		time.Sleep(600 * time.Millisecond)
		ctx.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest(http.MethodGet, "/slow", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "at least one trace file should be created for slow request")

	logOutput := logBuf.String()
	require.NotEmpty(t, logOutput, "log should not be empty")

	var logEntry LogEntry
	err = json.Unmarshal([]byte(logOutput), &logEntry)
	require.NoError(t, err, "log should be valid JSON")

	assert.Equal(t, "info", logEntry.Level)
	assert.Equal(t, http.MethodGet, logEntry.Method)
	assert.Equal(t, "/slow", logEntry.URL)
	assert.Equal(t, "/slow", logEntry.Path)
	assert.Equal(t, http.StatusOK, logEntry.RespStatus)
	assert.GreaterOrEqual(t, logEntry.ElapsedMs, 600*time.Millisecond, "elapsed time should be at least 600ms")
}

func TestRequestLogMiddlewareWithoutFlightRecorder(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	gin.SetMode(gin.TestMode)
	router := gin.New()

	lgr := logger.New("info", &logBuf)
	router.Use(middleware.RequestLogMiddleware(lgr, nil))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	logOutput := logBuf.String()
	require.NotEmpty(t, logOutput, "log should not be empty")

	var logEntry LogEntry
	err := json.Unmarshal([]byte(logOutput), &logEntry)
	require.NoError(t, err, "log should be valid JSON")

	assert.Equal(t, "info", logEntry.Level)
	assert.Equal(t, http.MethodGet, logEntry.Method)
	assert.Equal(t, "/test", logEntry.URL)
	assert.Equal(t, "/test", logEntry.Path)
	assert.Equal(t, http.StatusOK, logEntry.RespStatus)
}
