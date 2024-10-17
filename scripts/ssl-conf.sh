echo 'hostssl all all 0.0.0.0/0 cert clientcert=verify-full' > /var/lib/postgresql/data/pg_hba.conf
echo 'hostssl all all ::1/128 cert clientcert=verify-full' >> /var/lib/postgresql/data/pg_hba.conf
echo "hba_file = '/var/lib/postgresql/data/pg_hba.conf'" >> /var/lib/postgresql/data/postgresql.conf
echo 'ssl = on' >> /var/lib/postgresql/data/postgresql.conf
echo "ssl_ca_file = '/var/lib/postgresql/ca-cert.pem'" >> /var/lib/postgresql/data/postgresql.conf
echo "ssl_cert_file = '/var/lib/postgresql/server-cert.pem'" >> /var/lib/postgresql/data/postgresql.conf
echo "ssl_key_file = '/var/lib/postgresql/server-key.pem'" >> /var/lib/postgresql/data/postgresql.conf