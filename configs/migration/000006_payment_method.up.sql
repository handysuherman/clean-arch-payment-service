CREATE TABLE "payment_method" (
  "uid" varchar PRIMARY KEY NOT NULL,
  "payment_method_id" varchar NOT NULL,
  "payment_request_id" varchar,
  "payment_reference_id" varchar NOT NULL,
  "payment_business_id" varchar NOT NULL,
  "payment_customer_id" varchar NOT NULL,
  "payment_type" varchar NOT NULL,
  "payment_status" varchar NOT NULL,
  "payment_reusability" varchar NOT NULL,
  "payment_channel" varchar NOT NULL,
  "payment_amount" numeric(15,2),
  "payment_qr_code" text,
  "payment_virtual_account_number" varchar(255),
  "payment_url" text,
  "payment_description" text NOT NULL,
  "payment_failure_code" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT ('0001-01-01 00:00:00Z'),
  "expires_at" timestamptz NOT NULL DEFAULT ('0001-01-01 00:00:00Z'),
  "paid_at" timestamptz NOT NULL DEFAULT ('0001-01-01 00:00:00Z')
);

CREATE INDEX ON "payment_method" ("payment_customer_id");

CREATE INDEX ON "payment_method" ("payment_type");

CREATE INDEX ON "payment_method" ("payment_status");

CREATE INDEX ON "payment_method" ("payment_channel");

CREATE INDEX ON "payment_method" ("payment_reference_id");

CREATE INDEX ON "payment_method" ("payment_business_id");

CREATE INDEX ON "payment_method" ("payment_method_id");

CREATE INDEX ON "payment_method" ("payment_request_id");

CREATE INDEX ON "payment_method" ("payment_reusability");

CREATE INDEX ON "payment_method" ("payment_failure_code");

COMMENT ON COLUMN "payment_method"."payment_qr_code" IS 'for QR_CODE payment type';

COMMENT ON COLUMN "payment_method"."payment_virtual_account_number" IS 'for BANK payment type';

COMMENT ON COLUMN "payment_method"."payment_url" IS 'for EWALLET payment type';

COMMENT ON COLUMN "payment_method"."payment_description" IS 'LIMIT Text Should not exceeding 100 chars';

COMMENT ON COLUMN "payment_method"."payment_failure_code" IS 'See here https://docs.xendit.co/id/subscriptions-payment-failure-code for reference';

ALTER TABLE "payment_method" ADD FOREIGN KEY ("payment_reusability") REFERENCES "payment_reusability" ("prname");

ALTER TABLE "payment_method" ADD FOREIGN KEY ("payment_type") REFERENCES "payment_type" ("ptname");

ALTER TABLE "payment_method" ADD FOREIGN KEY ("payment_status") REFERENCES "payment_status" ("psname");

ALTER TABLE "payment_method" ADD FOREIGN KEY ("payment_channel") REFERENCES "payment_channel" ("pcname");