CREATE TABLE IF NOT EXISTS nodes (
  node_id UUID PRIMARY KEY,
  hostname VARCHAR(64) NOT NULL,    -- node hostname
  ip_address VARCHAR(64) NOT NULL,  -- node ip_address
  group_id INTEGER NOT NULL,        -- group_id the node belongs to
  vcpu INTEGER NOT NULL,            ---|
  memory INTEGER NOT NULL,          ---+-> node resource size
  storage INTEGER NOT NULL,         ---|

  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);

-- development purpose
-- INSERT INTO 
--   nodes(node_id, hostname, ip_address, group_id, vcpu, memory, storage)
-- VALUES
--   ('93069ae0-5706-4cf8-ad40-5fee0aedfee5', 'ajk-1', '10.21.73.100', 1, 16, 32, 128),
--   ('98eba7ae-cffa-43f9-bc0c-f1daaaca7a51', 'ajk-2', '10.21.73.101', 1, 16, 32, 64),
--   ('6b8f8c7c-ce30-4235-bd56-25acc926b8f0', 'ajk-3', '10.21.73.102', 1, 16, 32, 80);
