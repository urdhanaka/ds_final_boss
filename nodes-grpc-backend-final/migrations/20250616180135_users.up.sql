CREATE TABLE IF NOT EXISTS users (
  user_id UUID PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  group_id INTEGER NOT NULL,

  CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups (group_id)
);

INSERT INTO 
  users(user_id, name, email, password, group_id)
VALUES
  ('60872bf3-54ae-49a8-99c7-cd6de9e56c20', 'AJK-User-1', 'ajk-1@mail.com', 'ajkuser1', 1),
  ('45ff9081-1192-4db8-83e5-21f21225bbf8', 'AJK-User-2', 'ajk-2@mail.com', 'ajkuser2', 1),
  ('a9c49af3-f969-4ea8-b89d-6fef5989ed2a', 'KBJ-User-1', 'kbj-1@mail.com', 'kbjuser1', 3);
