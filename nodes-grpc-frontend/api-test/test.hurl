# POST http://localhost:3000/nodes
# {
#   "hostname": "host-1",
#   "ip_address": "69.69.69.69",
#   "grpc_port": ":1234"
# }
#
# POST http://localhost:3000/nodes
# {
#   "hostname": "host-2",
#   "ip_address": "69.69.69.70",
#   "grpc_port": ":1234"
# }

# fiberApp.Post("/create-cluster", nodeHandler.CreateCluster)
POST http://localhost:3000/create-cluster
# {
#   "token": "12345678"
# }

# GET http://localhost:3000/nodes

	# Hostname  string `json:"hostname"`
	# IpAddress string `json:"ip_address"`
