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
	clusterNodeRepository := repositories.NewClusterNodeRepository(dbPool)
	clusterRepository := repositories.NewClusterRepository(dbPool)
	nodeRepository := repositories.NewNodeRepository(dbPool)
	groupRepository := repositories.NewGroupRepository(dbPool)
	userRepository := repositories.NewUserRepository(dbPool)

	clusterService := services.NewClusterService(
		clusterNodeRepository,
		clusterRepository,
		nodeRepository,
		groupRepository,
	)
	redisQueueService := services.NewRedisJobQueue(client, clusterService)
	jwtService := services.NewJwtService()
	nodeService := services.NewNodeService(nodeRepository)
	userService := services.NewUserService(userRepository, groupRepository, jwtService)

	clusterHandler := handlers.NewClusterHandler(redisQueueService, jwtService, clusterService)
	userHandler := handlers.NewUserHandler(userService)
	nodeHandler := handlers.NewNodeHandler(nodeService, userService)
	otherHandler := handlers.NewOtherHandler()

	apiGroup := app.Group("/api")
	apiGroup.GET("/health", otherHandler.Checkhealth)

	clusterGroup := apiGroup.Group("/clusters")
	{
		clusterGroup.POST("", clusterHandler.CreateCluster)
		clusterGroup.GET("/user-clusters", handlers.Authenticate(jwtService), clusterHandler.GetUserCluster)
        clusterGroup.GET("/:cluster_id", handlers.Authenticate(jwtService), clusterHandler.GetClusterDetails)
	}

	userGroup := apiGroup.Group("/users")
	{
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/me", handlers.Authenticate(jwtService), userHandler.MeUser)
	}

	nodeGroup := apiGroup.Group("/nodes")
	{
		nodeGroup.GET("/group-nodes", handlers.Authenticate(jwtService), nodeHandler.GetGroupCluster)
	}
}
