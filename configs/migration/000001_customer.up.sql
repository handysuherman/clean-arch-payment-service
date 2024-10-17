CREATE TABLE "customer" (
  "uid" varchar PRIMARY KEY NOT NULL,
  "customer_app_id" varchar NOT NULL,
  "payment_customer_id" varchar NOT NULL,
  "customer_name" varchar(30) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "email" varchar,
  "phone_number" varchar
);

CREATE UNIQUE INDEX ON "customer" ("payment_customer_id");

CREATE UNIQUE INDEX ON "customer" ("customer_app_id");