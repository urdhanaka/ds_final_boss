package main

import (
	"context"
	"nodes-grpc-be/handlers"
	"nodes-grpc-be/repositories"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func setRouters(app *gin.Engine, dbPool *pgxpool.Pool, client *redis.Client) {
	clusterNodeRepository := repositories.NewClusterNodeRepository(dbPool)
	clusterRepository := repositories.NewClusterRepository(dbPool)
	nodeRepository := repositories.NewNodeRepository(dbPool)
	groupRepository := repositories.NewGroupRepository(dbPool)
	userRepository := repositories.NewUserRepository(dbPool)

	clusterService := services.NewClusterService(
		clusterNodeRepository,
		clusterRepository,
		userRepository,
		nodeRepository,
		groupRepository,
	)
	redisQueueService := services.NewRedisJobQueue(client, clusterService)
	redisQueueService.StartWorker(context.Background())

	jwtService := services.NewJwtService()
	nodeService := services.NewNodeService(nodeRepository, groupRepository)
	userService := services.NewUserService(userRepository, groupRepository, jwtService)

	clusterHandler := handlers.NewClusterHandler(redisQueueService, jwtService, clusterService, userService)
	userHandler := handlers.NewUserHandler(userService)
	nodeHandler := handlers.NewNodeHandler(nodeService, userService)
	otherHandler := handlers.NewOtherHandler()

	apiGroup := app.Group("/api")
	apiGroup.GET("/health", otherHandler.Checkhealth)

	clusterGroup := apiGroup.Group("/clusters")
	{
		clusterGroup.POST("", handlers.Authenticate(jwtService), clusterHandler.CreateCluster)
		clusterGroup.GET("", handlers.Authenticate(jwtService), clusterHandler.GetUserCluster)
		clusterGroup.GET("/:cluster_id", handlers.Authenticate(jwtService), clusterHandler.GetClusterDetails)
		clusterGroup.DELETE("/:cluster_id", handlers.Authenticate(jwtService), clusterHandler.GetClusterDetails)
	}

	userGroup := apiGroup.Group("/users")
	{
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/me", handlers.Authenticate(jwtService), userHandler.MeUser)
	}

	nodeGroup := apiGroup.Group("/nodes")
	{
		nodeGroup.POST("", nodeHandler.AddNode)
		nodeGroup.GET("/group-nodes", handlers.Authenticate(jwtService), nodeHandler.GetGroupCluster)
	}
}
