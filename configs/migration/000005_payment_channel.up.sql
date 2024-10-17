CREATE TABLE "payment_channel" (
  "uid" varchar PRIMARY KEY NOT NULL,
  "pcname" varchar NOT NULL,
  "pc_type" varchar NOT NULL,
  "logo_src" varchar NOT NULL DEFAULT 'logo',
  "min_amount" numeric(15,2) NOT NULL,
  "max_amount" numeric(15,2) NOT NULL,
  "tax" numeric(15,2) NOT NULL,
  "is_tax_percentage" boolean NOT NULL DEFAULT false,
  "is_active" boolean NOT NULL DEFAULT false,
  "is_available" boolean NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX ON "payment_channel" ("pcname");

INSERT INTO "public"."payment_channel" ("uid","pcname","logo_src", "pc_type","min_amount","max_amount","tax","is_tax_percentage","is_active","is_available") VALUES 
('01HY0AGZRV465F2BEZMQD46M3N','OVO','ovo','EWALLET',100.00,2000000.00,2.73,'TRUE','TRUE','TRUE'),
('01HY0AGZRYZ6SY3VW7V402YRM1','DANA','dana','EWALLET',100.00,2000000.00,1.50,'TRUE','TRUE','TRUE'),
('01HY0AGZS2R3G8MFQ96JR489VX','LINKAJA','link-aja','EWALLET',100.00,2000000.00,2.70,'TRUE','TRUE','TRUE'),
('01HY0AGZSZW4D54W3T58004EKP','ASTRAPAY','astra-pay','EWALLET',100.00,2000000.00,1.50,'TRUE','TRUE','TRUE'),
('01HY0AGZT8JD5QRYMW4YP2RZ9K','SHOPEEPAY','shopee-pay','EWALLET',100.00,2000000.00,4.00,'TRUE','TRUE','TRUE'),
('01HY0AGZTGM3GWCQM3GPMD6JKH','QRIS','qris','QR_CODE',1.00,10000000.00,0.70,'TRUE','TRUE','TRUE'),
('01HY0DJYQW0VADGD6WD2WDVJTR','BCA','bca','VIRTUAL_ACCOUNT',10000.00,50000000.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0DK2GN4SEPJN3859X24YKV','BNI','bni','VIRTUAL_ACCOUNT',1.00,50000000.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0DK67RW7FFS0210AZ3YK96','BRI','bri','VIRTUAL_ACCOUNT',1.00,50000000000.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0DKE7J9WKSTK1V7FM1YPJ6','CIMB','cimb','VIRTUAL_ACCOUNT',10000.00,50000000.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0DKHWT7WQ8F1YNNA3CMM7D','MANDIRI','mandiri','VIRTUAL_ACCOUNT',1.00,50000000000.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0EKQRBRPSCJ242MJ1YCVB0','PERMATA','permata','VIRTUAL_ACCOUNT',1.00,9999999999.00,4000.00,'FALSE','TRUE','TRUE'),
('01HY0DKP4V3Q0M8QVHHPKDY7CR','BSI','bsi','VIRTUAL_ACCOUNT',1.00,50000000000.00,4000.00,'FALSE','TRUE','TRUE');

ALTER TABLE "payment_channel" ADD FOREIGN KEY ("pc_type") REFERENCES "payment_type" ("ptname");