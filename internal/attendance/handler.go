package attendance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pli/absensi-api/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CheckIn(c *gin.Context) {
	claims := auth.GetClaims(c)

	latStr := c.PostForm("lat")
	lngStr := c.PostForm("lng")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat tidak valid"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lng tidak valid"})
		return
	}

	var photoFile interface{ Read([]byte) (int, error) } = nil
	file, header, _ := c.Request.FormFile("photo")

	rec, err := h.service.CheckIn(c.Request.Context(), claims.UserID, lat, lng, file, header)
	_ = photoFile
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rec)
}

func (h *Handler) CheckOut(c *gin.Context) {
	claims := auth.GetClaims(c)

	rec, err := h.service.CheckOut(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rec)
}

func (h *Handler) Today(c *gin.Context) {
	claims := auth.GetClaims(c)

	status, err := h.service.TodayStatus(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) History(c *gin.Context) {
	claims := auth.GetClaims(c)

	records, err := h.service.History(c.Request.Context(), claims.UserID,
		c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *Handler) AllAttendances(c *gin.Context) {
	records, err := h.service.AllToday(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}
