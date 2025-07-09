CREATE TABLE IF NOT EXISTS cluster_nodes (
  id SERIAL PRIMARY KEY,
  cluster_id UUID,  -- cluster
  node_id UUID,     -- node
  instance_name TEXT, -- domain name

  CONSTRAINT fk_cluster_id FOREIGN KEY (cluster_id) REFERENCES clusters (cluster_id) ON DELETE CASCADE,
  CONSTRAINT fk_node_id FOREIGN KEY (node_id) REFERENCES nodes (node_id)
);
