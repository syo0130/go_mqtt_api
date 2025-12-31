package http

import (
	"net/http"
	"strconv"

	"go_mqtt_api/internal/usecase"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests
type Handler struct {
	service *usecase.SensorDataService
}

// NewHandler creates a new HTTP handler
func NewHandler(service *usecase.SensorDataService) *Handler {
	return &Handler{
		service: service,
	}
}

// GetLatestValue handles GET /devices/:id/latest
func (h *Handler) GetLatestValue(c echo.Context) error {
	deviceID := c.Param("id")
	if deviceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device ID is required",
		})
	}

	latest, err := h.service.GetLatestValue(c.Request().Context(), deviceID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, latest)
}

// GetHistory handles GET /devices/:id/history
func (h *Handler) GetHistory(c echo.Context) error {
	deviceID := c.Param("id")
	if deviceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device ID is required",
		})
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 100 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	// Get history
	records, err := h.service.GetHistory(c.Request().Context(), deviceID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get total count for pagination
	total, err := h.service.GetHistoryCount(c.Request().Context(), deviceID)
	if err != nil {
		// Log error but don't fail the request
		total = int64(len(records))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  records,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/devices/:id/latest", h.GetLatestValue)
	e.GET("/devices/:id/history", h.GetHistory)
}

