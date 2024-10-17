CREATE TABLE "payment_reusability" (
  "prname" varchar PRIMARY KEY NOT NULL
);

INSERT INTO "public"."payment_reusability" ("prname") VALUES 
('MULTIPLE_USE'),
('ONE_TIME_USE');