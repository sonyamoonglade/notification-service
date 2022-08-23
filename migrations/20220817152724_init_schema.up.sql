
CREATE TABLE IF NOT EXISTS "events"(
    "event_id" INTEGER PRIMARY KEY UNIQUE NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "translate" varchar(255) UNIQUE NOT NULL
 );

CREATE TABLE IF NOT EXISTS "subscribers"(
    "subscriber_id" SERIAL PRIMARY KEY,
    "phone_number" varchar(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "subscriptions"(
    "subscription_id" SERIAL PRIMARY KEY,
    "event_id" INTEGER NOT NULL,
    "subscriber_id" INTEGER NOT NULL
);

ALTER TABLE "subscriptions" ADD CONSTRAINT "event_id_fk"
    FOREIGN KEY("event_id")
    REFERENCES events("event_id")
    ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD CONSTRAINT "subscriber_id_fk"
    FOREIGN KEY("subscriber_id")
    REFERENCES subscribers("subscriber_id")
    ON DELETE CASCADE;

-- Make sure on database level that one subscriber cant subscribe to the same event multiple times.
ALTER TABLE "subscriptions" ADD CONSTRAINT "event_id_sub_id_unique"
    UNIQUE("event_id","subscriber_id");

CREATE TABLE IF NOT EXISTS "telegram_subscribers"(
    "subscriber_id" INTEGER NOT NULL,
    "telegram_id" BIGINT NOT NULL
);

-- Make sure that one telegram user is one subscriber only
ALTER TABLE "telegram_subscribers" ADD CONSTRAINT "teleg_id_unqiue"
    UNIQUE("subscriber_id");

ALTER TABLE "telegram_subscribers" ADD CONSTRAINT "subscriber_id_unique"
    UNIQUE("telegram_id");