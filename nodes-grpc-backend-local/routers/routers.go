package routers

import (
	"nodes-grpc-backend-local/handlers"
	"nodes-grpc-backend-local/repository"
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetRouters(app *gin.Engine, dbPool *pgxpool.Pool, redis *redis.Client) {
	nodeRepository := repository.NewNodeRepository(dbPool)
	userRepository := repository.NewUserRepository(dbPool)
	groupRepository := repository.NewGroupRepository(dbPool)
	clusterRepository := repository.NewClusterRepository(dbPool)
	clusterNodeRepository := repository.NewClusterNodeRepository(dbPool)

	nodeService := services.NewNodeService(nodeRepository, groupRepository)
	userService := services.NewUserService(userRepository)
	groupService := services.NewGroupService(groupRepository)
	clusterService := services.NewClusterService(clusterRepository, nodeRepository, clusterNodeRepository)
	queueService := services.NewQueueService(redis)
	jwtService := services.NewJwtService()

	nodeHandler := handlers.NewNodeHandler(nodeService)
	userHandler := handlers.NewUserHandler(userService, jwtService)
	groupHandler := handlers.NewGroupHandler(groupService)
	clusterHandler := handlers.NewClusterHandler(clusterService, queueService)

	apiGroup := app.Group("/api")
	v1 := apiGroup.Group("/v1")

	usersGroup := v1.Group("/users")
	{
		usersGroup.GET("", userHandler.GetAll)
		usersGroup.POST("/login", userHandler.Login)
	}

	groupsGroup := v1.Group("/groups")
	{
		groupsGroup.GET("", groupHandler.GetAllGroups)
	}

	nodesGroup := v1.Group("/nodes")
	{
		nodesGroup.GET("", nodeHandler.GetAllNodes)
		nodesGroup.POST("", nodeHandler.AddNode)
		// nodesGroup.POST("/assign-nodes", )
		// nodesGroup.GET("/status")
	}

	clusterGroup := v1.Group("/clusters")
	{
		clusterGroup.POST("", clusterHandler.CreateCluster)
	}
}
