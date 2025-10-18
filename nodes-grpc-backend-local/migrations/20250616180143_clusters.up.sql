CREATE TABLE IF NOT EXISTS clusters (
  cluster_id UUID PRIMARY KEY,
  cluster_name VARCHAR(128) NOT NULL, -- cluster_name
  user_id UUID NOT NULL,              -- the user_id that created the cluster
  group_id INTEGER NOT NULL,          -- the group_id the cluster belongs to
  created_at TIMESTAMP NOT NULL,      -- timestamp created_at

  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (user_id),
  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);
