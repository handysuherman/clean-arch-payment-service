CREATE TABLE "payment_status" (
  "psname" varchar PRIMARY KEY NOT NULL
);

INSERT INTO "public"."payment_status" ("psname") VALUES 
('ACTIVE'),
('INACTIVE'),
('PENDING'),
('EXPIRED'),
('FAILED'),
('REQUIRES_ACTION'),
('CANCELED'),
('SUCCEEDED'),
('VOIDED'),
('AWAITING_CAPTURE');