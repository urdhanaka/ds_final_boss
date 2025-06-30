CREATE TABLE IF NOT EXISTS groups (
  group_id INTEGER PRIMARY KEY,
  name VARCHAR(128) NOT NULL,       -- lab name
  vcpu INTEGER NOT NULL,            ---|
  memory INTEGER NOT NULL,          ---+> All of this is static
  storage INTEGER NOT NULL,         ---|
  node_size INTEGER NOT NULL,       -- static size of the node on a cluster
  current_cluster INTEGER NOT NULL, -- current cluster that exists
  max_cluster INTEGER NOT NULL      -- maximum size of the cluster 
                                    -- that can be created by this group
);

INSERT INTO 
  groups(group_id, name, vcpu, memory, storage, node_size, current_cluster, max_cluster)
VALUES
-- id, name, vcpu, memory, storage, node_size, current_cluster max_cluster
  (1,  'AJK', 2,   2,      16,      2,         0,              4),
  (2,  'RPL', 2,   4,      16,      2,         0,              2),
  (3,  'KBJ', 4,   4,      16,      2,         0,              4);
