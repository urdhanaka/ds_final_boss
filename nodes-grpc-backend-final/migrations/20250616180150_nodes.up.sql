CREATE TABLE IF NOT EXISTS nodes (
  node_id UUID PRIMARY KEY,
  hostname VARCHAR(128) NOT NULL,   -- node hostname
  ip_address VARCHAR(64) NOT NULL,  -- node ip_address
  group_id INTEGER NOT NULL,        -- group_id the node belongs to
  cpu INTEGER NOT NULL,             ---|
  ram INTEGER NOT NULL,             ---+> node resource size
  storage INTEGER NOT NULL,         ---|

  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);
