CREATE TABLE IF NOT EXISTS groups (
  group_id INTEGER PRIMARY KEY,
  name VARCHAR(128) NOT NULL,   -- lab name
  vcpu INTEGER NOT NULL,        ---|
  ram INTEGER NOT NULL,         ---+> All of this is static
  storage INTEGER NOT NULL,     ---|
  node_size INTEGER NOT NULL,   -- static size of the node on a cluster
  max_cluster INTEGER NOT NULL  -- maximum size of the cluster 
                                -- that can be created by this group
);

INSERT INTO 
  groups(group_id, name, vcpu, ram, storage, node_size, max_cluster)
VALUES
-- id, name, vcpu, ram, storage, node_size, max_cluster
  (1,  'AJK', 4,   4,   32,      3,         4),
  (2,  'RPL', 2,   4,   16,      2,         2),
  (3,  'KBJ', 4,   4,   32,      3,         4);
