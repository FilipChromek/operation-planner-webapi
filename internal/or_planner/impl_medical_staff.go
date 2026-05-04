package or_planner

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/FilipChromek/operation-planner-webapi/internal/db_service"
)

type implMedicalStaffAPI struct{}

func NewMedicalStaffApi() MedicalStaffAPI {
	return &implMedicalStaffAPI{}
}

func staffDb(c *gin.Context) (db_service.DbService[MedicalStaff], bool) {
	v, ok := c.Get("staff_db")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "staff_db not in context"})
		return nil, false
	}
	db, ok := v.(db_service.DbService[MedicalStaff])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "staff_db has wrong type"})
		return nil, false
	}
	return db, true
}

func (o *implMedicalStaffAPI) GetStaff(c *gin.Context) {
	db, ok := staffDb(c)
	if !ok {
		return
	}
	list, err := db.FindAll(c)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (o *implMedicalStaffAPI) GetStaffMember(c *gin.Context) {
	db, ok := staffDb(c)
	if !ok {
		return
	}
	id := c.Param("staffId")
	s, err := db.FindDocument(c, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "staff member not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implMedicalStaffAPI) CreateStaff(c *gin.Context) {
	db, ok := staffDb(c)
	if !ok {
		return
	}
	var s MedicalStaff
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.Id == "" {
		s.Id = uuid.NewString()
	}
	err := db.CreateDocument(c, s.Id, &s)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "staff already exists"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implMedicalStaffAPI) UpdateStaffMember(c *gin.Context) {
	db, ok := staffDb(c)
	if !ok {
		return
	}
	id := c.Param("staffId")
	var s MedicalStaff
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.Id = id
	err := db.UpdateDocument(c, id, &s)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "staff member not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}
