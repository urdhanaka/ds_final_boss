CREATE TABLE IF NOT EXISTS clusters (
  cluster_id UUID PRIMARY KEY,
  cluster_name VARCHAR(128) NOT NULL,   -- cluster_name
  user_id UUID NOT NULL,                -- the user_id that created the cluster
  group_id INTEGER NOT NULL,            -- the group_id the cluster belongs to
  cluster_status VARCHAR(16) NOT NULL,  -- creation status
  ip_address VARCHAR(32),               -- ip address to access the cluster dashboard
  access_token VARCHAR(1024),           -- token to access the kubernetes cluster
  kubeconfig_contents BYTEA,            -- the kubeconfig file bytes
  created_at TIMESTAMP NOT NULL,        -- timestamp created_at

  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (user_id),
  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);

-- development purpose
-- INSERT INTO
--   clusters(cluster_id, cluster_name, user_id, group_id, cluster_status, ip_address, access_token, created_at)
-- VALUES
--   ('418f7bc3-0de7-43f8-b968-40d821d822e7', 'vc-1', '60872bf3-54ae-49a8-99c7-cd6de9e56c20', 1, 'created', '192.168.1.1', 'acclakjdklasjdlkabsklasf', '2004-10-19 10:23:54'),
--   ('03793870-185f-4ed8-9c23-424372e1112c', 'vc-2', '60872bf3-54ae-49a8-99c7-cd6de9e56c20', 1, 'creating', '192.168.1.23', 'hshfalksbflabsklasf', '2004-10-19 10:23:54');
  -- ('51425c35-6428-4996-880b-6396c2e31e84', 'vc-1', 2);
