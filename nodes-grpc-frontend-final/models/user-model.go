package models

import "github.com/google/uuid"

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
