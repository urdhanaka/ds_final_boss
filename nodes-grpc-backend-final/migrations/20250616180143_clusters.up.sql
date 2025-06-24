CREATE TABLE IF NOT EXISTS clusters (
  cluster_id UUID PRIMARY KEY,
  cluster_name VARCHAR(128) NOT NULL,   -- cluster_name
  user_id UUID NOT NULL,                -- the user_id that created the cluster
  group_id INTEGER NOT NULL,            -- the group_id the cluster belongs to
  cluster_status VARCHAR(16) NOT NULL,  -- creation status
  ip_address VARCHAR(32),               -- ip address to access the cluster dashboard
  access_token VARCHAR(256),            -- token to access the kubernetes cluster
  created_at TIMESTAMP NOT NULL,        -- timestamp created_at

  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (user_id),
  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);
