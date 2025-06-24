package models

import "github.com/google/uuid"

type AddUser struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
	GroupId  int    `json:"group_id" binding:"required"`
}

type LoginUser struct {
	Email    string `json:"email"       binding:"required"`
	Password string `json:"password"    binding:"required"`
}

type MeUserReturn struct {
	UserId         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Group          string    `json:"group"`
	Vcpu           int       `json:"vcpu"`
	Ram            int       `json:"ram"`
	Storage        int       `json:"storage"`
	NodeSize       int       `json:"node_size"`
	CurrentCluster int       `json:"current_cluster"`
	MaxCluster     int       `json:"max_cluster"`
}
