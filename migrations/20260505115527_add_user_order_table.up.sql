CREATE TABLE user_order
(
    id         UUID           NOT NULL PRIMARY KEY,
    user_id    UUID           NOT NULL,
    order_id   VARCHAR(255)   NOT NULL,
    status     SMALLINT       NOT NULL,
    accrual    NUMERIC(15, 2) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE INDEX user_order_user_id_idx ON user_order (user_id);
CREATE INDEX user_order_order_id_idx ON user_order (order_id);
CREATE INDEX user_order_status_idx ON user_order (status);

