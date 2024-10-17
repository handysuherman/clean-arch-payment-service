CREATE TABLE "payment_type" (
  "ptname" varchar PRIMARY KEY NOT NULL
);

INSERT INTO "public"."payment_type" ("ptname") VALUES 
('CARD'),
('CRYPTOCURRENCY'),
('DIRECT_BANK_TRANSFER'),
('DIRECT_DEBIT'),
('EWALLET'),
('OVER_THE_COUNTER'),
('QR_CODE'),
('VIRTUAL_ACCOUNT');