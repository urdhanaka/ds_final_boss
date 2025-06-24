package main

import (
	"nodes-grpc-be/handlers"
	"nodes-grpc-be/repositories"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func setRouters(app *gin.Engine, dbPool *pgxpool.Pool, client *redis.Client) {
	// clusterNodeRepository := repositories.NewClusterNodeRepository(dbPool)
	// clusterRepository := repositories.NewClusterRepository(dbPool)
	// nodeRepository := repositories.NewNodeRepository(dbPool)
	groupRepository := repositories.NewGroupRepository(dbPool)
	userRepository := repositories.NewUserRepository(dbPool)

	// redisQueueService := services.NewRedisJobQueue(client)
	// clusterService := services.NewClusterService(
	// 	clusterNodeRepository,
	// 	clusterRepository,
	// 	nodeRepository,
	// 	groupRepository,
	// )
	jwtService := services.NewJwtService()
	userService := services.NewUserService(userRepository, groupRepository, jwtService)

	// clusterHandler := handlers.NewClusterHandler(redisQueueService)
	userHandler := handlers.NewUserHandler(userService)

	apiGroup := app.Group("/api")

	// clusterGroup := apiGroup.Group("/clusters")
	// {
	// 	clusterGroup.POST("", clusterHandler.CreateCluster)
	// }

	userGroup := apiGroup.Group("/users")
	{
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/me", handlers.Authenticate(jwtService), userHandler.MeUser)
	}
}
