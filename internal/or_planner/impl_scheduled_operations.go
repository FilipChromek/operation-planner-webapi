package or_planner

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/FilipChromek/operation-planner-webapi/internal/db_service"
)

type implScheduledOperationsAPI struct{}

func NewScheduledOperationsApi() ScheduledOperationsAPI {
	return &implScheduledOperationsAPI{}
}

// reconcileSchedule sorts operations and resolves overlaps (shifts later ones).
func reconcileSchedule(ops []ScheduledOperation) []ScheduledOperation {
	if len(ops) <= 1 {
		return ops
	}
	slices.SortFunc(ops, func(a, b ScheduledOperation) int {
		if a.ScheduledStart.Before(b.ScheduledStart) {
			return -1
		}
		if a.ScheduledStart.After(b.ScheduledStart) {
			return 1
		}
		return 0
	})
	for i := 1; i < len(ops); i++ {
		prev := ops[i-1]
		prevEnd := prev.ScheduledStart.Add(time.Duration(prev.DurationMinutes) * time.Minute)
		if ops[i].ScheduledStart.Before(prevEnd) {
			ops[i].ScheduledStart = prevEnd
		}
	}
	return ops
}

func (o *implScheduledOperationsAPI) GetRoomOperations(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	roomId := c.Param("roomId")
	room, err := db.FindDocument(c, roomId)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	ops := room.ScheduledOperations
	if ops == nil {
		ops = []ScheduledOperation{}
	}
	c.JSON(http.StatusOK, ops)
}

func (o *implScheduledOperationsAPI) GetOperation(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	roomId := c.Param("roomId")
	opId := c.Param("opId")
	room, err := db.FindDocument(c, roomId)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	idx := slices.IndexFunc(room.ScheduledOperations, func(op ScheduledOperation) bool {
		return op.Id == opId
	})
	if idx < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "operation not found"})
		return
	}
	c.JSON(http.StatusOK, room.ScheduledOperations[idx])
}

func (o *implScheduledOperationsAPI) CreateOperation(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	roomId := c.Param("roomId")
	var op ScheduledOperation
	if err := c.BindJSON(&op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if op.PatientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patientId is required"})
		return
	}
	if op.Id == "" || op.Id == "@new" {
		op.Id = uuid.NewString()
	}
	room, err := db.FindDocument(c, roomId)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if slices.IndexFunc(room.ScheduledOperations, func(o ScheduledOperation) bool {
		return o.Id == op.Id
	}) >= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "operation with such id exists"})
		return
	}
	room.ScheduledOperations = append(room.ScheduledOperations, op)
	room.ScheduledOperations = reconcileSchedule(room.ScheduledOperations)
	if err := db.UpdateDocument(c, roomId, room); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	idx := slices.IndexFunc(room.ScheduledOperations, func(o ScheduledOperation) bool {
		return o.Id == op.Id
	})
	c.JSON(http.StatusOK, room.ScheduledOperations[idx])
}

func (o *implScheduledOperationsAPI) UpdateOperation(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	roomId := c.Param("roomId")
	opId := c.Param("opId")
	var update ScheduledOperation
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room, err := db.FindDocument(c, roomId)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	idx := slices.IndexFunc(room.ScheduledOperations, func(o ScheduledOperation) bool {
		return o.Id == opId
	})
	if idx < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "operation not found"})
		return
	}
	update.Id = opId
	room.ScheduledOperations[idx] = update
	room.ScheduledOperations = reconcileSchedule(room.ScheduledOperations)
	if err := db.UpdateDocument(c, roomId, room); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	idx = slices.IndexFunc(room.ScheduledOperations, func(o ScheduledOperation) bool {
		return o.Id == opId
	})
	c.JSON(http.StatusOK, room.ScheduledOperations[idx])
}

func (o *implScheduledOperationsAPI) DeleteOperation(c *gin.Context) {
	db, ok := roomsDb(c)
	if !ok {
		return
	}
	roomId := c.Param("roomId")
	opId := c.Param("opId")
	room, err := db.FindDocument(c, roomId)
	if err == db_service.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	idx := slices.IndexFunc(room.ScheduledOperations, func(o ScheduledOperation) bool {
		return o.Id == opId
	})
	if idx < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "operation not found"})
		return
	}
	room.ScheduledOperations = append(room.ScheduledOperations[:idx], room.ScheduledOperations[idx+1:]...)
	if err := db.UpdateDocument(c, roomId, room); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
