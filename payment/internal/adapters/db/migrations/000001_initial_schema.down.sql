-- payment/internal/adapters/db/migrations/000001_initial_schema.down.sql
DROP TABLE IF EXISTS outbox_messages;
DROP TABLE IF EXISTS payments;
