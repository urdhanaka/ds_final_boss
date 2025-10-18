CREATE TABLE IF NOT EXISTS users (
  user_id UUID PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  group_id INTEGER,           -- if superadmin, this can be null
  role VARCHAR(16) NOT NULL,  -- superadmin, admin, user

  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);
