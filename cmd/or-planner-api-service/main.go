package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/FilipChromek/operation-planner-webapi/api"
	"github.com/FilipChromek/operation-planner-webapi/internal/db_service"
	"github.com/FilipChromek/operation-planner-webapi/internal/or_planner"
)

func main() {
	log.Printf("Server started")
	port := os.Getenv("OR_PLANNER_API_PORT")
	if port == "" {
		port = "8080"
	}
	environment := os.Getenv("OR_PLANNER_API_ENVIRONMENT")
	if !strings.EqualFold(environment, "production") {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	corsMw := cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(corsMw)

	// Three separate Mongo collections (rooms, patients, staff)
	roomsDb := db_service.NewMongoService[or_planner.OperatingRoom](db_service.MongoServiceConfig{
		Collection: envOr("OR_PLANNER_API_MONGODB_COLLECTION_ROOMS", "rooms"),
	})
	patientsDb := db_service.NewMongoService[or_planner.Patient](db_service.MongoServiceConfig{
		Collection: envOr("OR_PLANNER_API_MONGODB_COLLECTION_PATIENTS", "patients"),
	})
	staffDb := db_service.NewMongoService[or_planner.MedicalStaff](db_service.MongoServiceConfig{
		Collection: envOr("OR_PLANNER_API_MONGODB_COLLECTION_STAFF", "staff"),
	})
	defer roomsDb.Disconnect(context.Background())
	defer patientsDb.Disconnect(context.Background())
	defer staffDb.Disconnect(context.Background())

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("rooms_db", roomsDb)
		ctx.Set("patients_db", patientsDb)
		ctx.Set("staff_db", staffDb)
		ctx.Next()
	})

	// Register routes from generated router
	handle := or_planner.ApiHandleFunctions{
		OperatingRoomsAPI:       or_planner.NewOperatingRoomsApi(),
		ScheduledOperationsAPI:  or_planner.NewScheduledOperationsApi(),
		PatientsAPI:             or_planner.NewPatientsApi(),
		MedicalStaffAPI:         or_planner.NewMedicalStaffApi(),
	}
	or_planner.NewRouterWithGinEngine(engine, handle)

	engine.GET("/openapi", api.HandleOpenApi)

	if err := engine.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}
