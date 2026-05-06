CREATE TABLE user_balance
(
    user_id       UUID           NOT NULL PRIMARY KEY,
    balance       NUMERIC(15, 2) NOT NULL  DEFAULT 0,
    withdrawn_sum NUMERIC(15, 2) NOT NULL  DEFAULT 0,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

