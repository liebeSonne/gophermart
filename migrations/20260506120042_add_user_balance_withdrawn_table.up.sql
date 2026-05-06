CREATE TABLE user_balance_withdrawn
(
    id         UUID           NOT NULL PRIMARY KEY,
    user_id    UUID           NOT NULL,
    order_id   VARCHAR(255)   NOT NULL,
    amount     NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

CREATE INDEX user_balance_withdrawn_user_id_idx ON user_balance_withdrawn (user_id);
CREATE INDEX user_balance_withdrawn_order_id_idx ON user_balance_withdrawn (order_id);