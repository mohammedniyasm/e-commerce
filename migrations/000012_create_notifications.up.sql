CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL,
    entity_type VARCHAR(50),
    entity_id VARCHAR(100),
    event_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT notifications_priority_check
        CHECK (priority IN ('info', 'medium', 'high', 'critical'))
);

CREATE INDEX idx_notifications_type
    ON notifications (type);

CREATE INDEX idx_notifications_priority
    ON notifications (priority);

CREATE INDEX idx_notifications_created_at
    ON notifications (created_at DESC);


CREATE TABLE notification_recipients (
    id BIGSERIAL PRIMARY KEY,
    notification_id UUID NOT NULL,
    admin_id BIGINT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    CONSTRAINT fk_notification_recipients_notification
        FOREIGN KEY (notification_id)
        REFERENCES notifications (id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_notification_recipients_admin
        FOREIGN KEY (admin_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_notification_recipient
        UNIQUE (notification_id, admin_id),

    CONSTRAINT notification_read_status_check
        CHECK (
            (is_read = FALSE AND read_at IS NULL)
            OR
            (is_read = TRUE AND read_at IS NOT NULL)
        )
);

CREATE INDEX idx_notification_recipients_admin_read
    ON notification_recipients (admin_id, is_read);

CREATE INDEX idx_notification_recipients_notification_id
    ON notification_recipients (notification_id);
