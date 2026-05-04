package or_planner

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/FilipChromek/operation-planner-webapi/internal/db_service"
)

type implPatientsAPI struct{}

func NewPatientsApi() PatientsAPI {
	return &implPatientsAPI{}
}

func patientsDb(c *gin.Context) (db_service.DbService[Patient], bool) {
	v, ok := c.Get("patients_db")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "patients_db not in context"})
		return nil, false
	}
	db, ok := v.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "patients_db has wrong type"})
		return nil, false
	}
	return db, true
}

func (o *implPatientsAPI) GetPatients(c *gin.Context) {
	db, ok := patientsDb(c)
	if !ok {
		return
	}
	patients, err := db.FindAll(c)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, patients)
}

func (o *implPatientsAPI) GetPatient(c *gin.Context) {
	db, ok := patientsDb(c)
	if !ok {
		return
	}
	id := c.Param("patientId")
	p, err := db.FindDocument(c, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, p)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implPatientsAPI) CreatePatient(c *gin.Context) {
	db, ok := patientsDb(c)
	if !ok {
		return
	}
	var p Patient
	if err := c.BindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p.Id == "" {
		p.Id = uuid.NewString()
	}
	err := db.CreateDocument(c, p.Id, &p)
	switch err {
	case nil:
		c.JSON(http.StatusOK, p)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "patient already exists"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implPatientsAPI) UpdatePatient(c *gin.Context) {
	db, ok := patientsDb(c)
	if !ok {
		return
	}
	id := c.Param("patientId")
	var p Patient
	if err := c.BindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p.Id = id
	err := db.UpdateDocument(c, id, &p)
	switch err {
	case nil:
		c.JSON(http.StatusOK, p)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}
