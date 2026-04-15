package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rameshsunkara/go-rest-api-example/internal/errors"
	"github.com/rameshsunkara/go-rest-api-example/internal/models/external"
	"github.com/rameshsunkara/go-rest-api-example/pkg/logger"
)

const ErrorContextKey = "app_error"

func ErrorHandlerMiddleware(lgr logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		errVal, exists := c.Get(ErrorContextKey)
		if !exists {
			return
		}

		err, ok := errVal.(error)
		if !ok {
			return
		}

		appErr, isAppErr := errors.IsAppError(err)
		if !isAppErr {
			appErr = errors.NewInternalError(
				"",
				errors.UnexpectedErrorMessage,
				"",
				err,
			)
		}

		l, _ := lgr.WithReqID(c)
		event := l.Error().
			Int("HttpStatusCode", appErr.StatusCode).
			Str("ErrorCode", appErr.ErrorCode)
		if appErr.Err != nil {
			event = event.Err(appErr.Err)
		}
		event.Msg(appErr.Message)

		apiErr := &external.APIError{
			HTTPStatusCode: appErr.StatusCode,
			ErrorCode:      appErr.ErrorCode,
			Message:        appErr.Message,
			DebugID:        appErr.DebugID,
		}

		c.AbortWithStatusJSON(appErr.StatusCode, apiErr)
	}
}

func HandleError(c *gin.Context, err error) {
	c.Set(ErrorContextKey, err)
}

func HandleErrorWithStatus(c *gin.Context, statusCode int, err error) {
	appErr, isAppErr := errors.IsAppError(err)
	if isAppErr {
		appErr.StatusCode = statusCode
		c.Set(ErrorContextKey, appErr)
	} else {
		c.Set(ErrorContextKey, &errors.AppError{
			ErrType:    errors.ErrorTypeInternal,
			StatusCode: statusCode,
			Message:    err.Error(),
			Err:        err,
		})
	}
}

func HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	if data == nil {
		c.Status(statusCode)
	} else {
		c.JSON(statusCode, data)
	}
}
