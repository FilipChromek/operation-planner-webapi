package or_planner

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/FilipChromek/operation-planner-webapi/internal/db_service"
)

type implOperatingRoomsAPI struct{}

func NewOperatingRoomsApi() OperatingRoomsAPI {
	return &implOperatingRoomsAPI{}
}

func roomsDb(c *gin.Context) (db_service.DbService[OperatingRoom], bool) {
	v, ok := c.Get("rooms_db")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rooms_db not in context"})
		return nil, false
	}
	db, ok := v.(db_service.DbService[OperatingRoom])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rooms_db has wrong type"})
		return nil, false
	}
	return db, true
}

func (o *implOperatingRoomsAPI) GetRooms(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	rooms, err := db.FindAll(c)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (o *implOperatingRoomsAPI) GetRoom(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	id := c.Param("roomId")
	room, err := db.FindDocument(c, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, room)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implOperatingRoomsAPI) CreateRoom(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	var room OperatingRoom
	if err := c.BindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if room.Id == "" {
		room.Id = uuid.NewString()
	}
	if room.ScheduledOperations == nil {
		room.ScheduledOperations = []ScheduledOperation{}
	}
	if room.PredefinedProcedures == nil {
		room.PredefinedProcedures = []ProcedureType{}
	}
	err := db.CreateDocument(c, room.Id, &room)
	switch err {
	case nil:
		c.JSON(http.StatusOK, room)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "room already exists"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func (o *implOperatingRoomsAPI) UpdateRoom(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	id := c.Param("roomId")
	existing, err := db.FindDocument(c, id)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	var update OperatingRoom
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// preserve scheduledOperations from existing if not provided
	if update.ScheduledOperations == nil {
		update.ScheduledOperations = existing.ScheduledOperations
	}
	update.Id = id
	if err := db.UpdateDocument(c, id, &update); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, update)
}
