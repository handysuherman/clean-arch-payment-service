PQSQL_ID=81d6fa93d9fe88c6148c-payment
JENKINS_IP=192.168.18.239
FOLDER=wisata-desa-payment-backend
RSYNC_OPTIONS := -avzhP --exclude='*_creds' --exclude='*.pkcs12' --exclude='*.p12' --exclude='*.cnf' --exclude='*-csr.pem' -e 'ssh'
CLEAN_PATH=./scripts/certs/production
DOCKER_REGISTRY_ENDPOINT=docker-registry.local:24115
DOCKER_IMAGE_NAME=wisata-desa-payment-backend

CERT_DIR=./tls
DN_NAME=payment_db
DB_SSL_CA_PATH=./tls/pqsql-copy/ca-cert.pem
DB_SSL_CLIENT_CERT_PATH=./tls/pqsql-copy/client-cert.pem
DB_SSL_CLIENT_KEY_PATH=./tls/pqsql-copy/client-key.pem
DB_PASSWORD=secret
DB_PORT=5432
DB_URL=postgresql://81d6fa93d9fe88c6148c-payment-pqsql-c-client:${DB_PASSWORD}@constantinopel.81d6fa93d9fe88c6148c-payment-pqsql-s.server:${DB_PORT}/${DN_NAME}?sslmode=require&sslrootcert=${DB_SSL_CA_PATH}&sslcert=${DB_SSL_CLIENT_CERT_PATH}&sslkey=${DB_SSL_CLIENT_KEY_PATH}
TEST_DB_URL=postgresql://root:secret@mock-pq-server:50241/mock_db?sslmode=disable

MIGRATION_PATH=configs/migration

.PHONY: setup
setup: versioning husky-setup

.PHONY: versioning
versioning:
	yarn add -D husky release-it @commitlint/cli @commitlint/config-conventional @release-it/conventional-changelog

.PHONY: husky-setup
husky-setup:
	./node_modules/husky/bin.mjs install && { echo '#!/bin/sh' && echo '' && echo '. "$$(dirname "$$0")/_/husky.sh"' && echo 'npx commitlint --edit'; } > .husky/commit-msg && chmod a+x .husky/commit-msg

.PHONY: updatecfg
updatecfg:
	etconf --config-file=configs/etcd/e-dev-config-cli.yaml

.PHONY: cmigrate
cmigrate:
	migrate create -ext sql -dir ${MIGRATION_PATH}/ -seq ${MIGRATE_NAME}

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: mock
mock:
	mockgen -package mock -destination internal/payment/repository/mock/mock.go -source=internal/payment/repository/repository.go
	mockgen -package wkmock -destination internal/payment/worker/mock/mock.go -source=internal/payment/domain/domain.go

.PHONY: docker
docker:
	docker build --no-cache -t ${DOCKER_REGISTRY_ENDPOINT}/${DOCKER_IMAGE_NAME}:latest -f build/docker/Dockerfile .
	docker push ${DOCKER_REGISTRY_ENDPOINT}/${DOCKER_IMAGE_NAME}:latest

.PHONY: test-updatecfg
test-updatecfg:
	~/go/bin/etconf --config-file=configs/etcd/e-test-config-cli.yaml

.PHONY: test-migrate
test-migrate:
	migrate -path ${MIGRATION_PATH} -database "$(TEST_DB_URL)" -verbose up

.PHONY: test-app
test-app:
	CGO_ENABLED=1 gotestsum \
	--jsonfile test-output.out \
	--format testdox \
	--packages="./internal/payment/usecase" \
	--packages="./internal/payment/repository" \
	--packages="./internal/payment/worker" \
	-- -cover -count=1 -race
	tparse -all -file=test-output.out

.PHONY: proto-app
proto-app:
	rm -rf internal/pb/*.go
	protoc --proto_path=internal/proto/app --go_out=internal/pb --go_opt=paths=source_relative \
	--go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative \
	internal/proto/app/*.proto

.PHONY: proto-kafka
proto-kafka:
	cd internal/proto/kafka && protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. kafka.proto

.PHONY: gen-gql
gen-gql:
	go run github.com/99designs/gqlgen generate

.PHONY: pq-ssl-mode
pq-ssl-mode:
	docker container exec -it ${FOLDER}-pqsql-1 bash /usr/local/bin/ssl-conf.sh
	docker container restart ${FOLDER}-pqsql-1

.PHONY: server
server:
	go run ./cmd/main.go --config-file=./etcd-config.yaml \
	--etcd-timeout=15s \
	--env="develop" \
	--hostname=192.168.18.160 \
	--uid=3f55729eac14a55c9051ef7a \
	--published-port=50050 \
	--health-check-addr=192.168.18.160 \
	--health-check-published-port=6880 

# ================================ Develop ================================

define get_cert
	mkdir -p $(1)
	rm -rf $(1)/*
	rsync $(RSYNC_OPTIONS) jenkins@${JENKINS_IP}:$(2)/service/client/ $(1)
	rsync $(RSYNC_OPTIONS) jenkins@${JENKINS_IP}:$(2)/service/server/ $(1)
	rsync $(RSYNC_OPTIONS) jenkins@${JENKINS_IP}:$(2)/service/ca/ca-cert.pem $(1)
endef

.PHONY: dev-migrateup
dev-migrateup:
	migrate -path ${MIGRATION_PATH} -database "$(DB_URL)" -verbose up

.PHONY: dev-migratedown
dev-migratedown:
	migrate -path ${MIGRATION_PATH} -database "$(DB_URL)" -verbose down -all

.PHONY: dev-cert
dev-cert: dev-etcd-cert dev-kafka-cert dev-pqsql-cert dev-consul-cert dev-redis-cert dev-pqsql-cert-copy

.PHONY: dev-etcd-cert
dev-etcd-cert:
	$(call get_cert,${CERT_DIR}/etcd,/home/jenkins/certs/production/wikanproductions/etcd/wikan)

.PHONY: dev-kafka-cert
dev-kafka-cert:
	$(call get_cert,./tls/kafka,~/certs/production/wikan-general/kafka)

.PHONY: dev-pqsql-cert
dev-pqsql-cert:
	$(call get_cert,./tls/pqsql,~/certs/production/wikanproductions/pqsql/${PQSQL_ID})

.PHONY: dev-consul-cert
dev-consul-cert:
	$(call get_cert,./tls/consul,~/certs/production/wikanproductions/consul/${PQSQL_ID})

.PHONY: dev-pqsql-cert-copy
dev-pqsql-cert-copy:
	$(call get_cert,./tls/pqsql-copy,~/certs/production/wikanproductions/pqsql/${PQSQL_ID})
	
.PHONY: dev-redis-cert
dev-redis-cert:
	$(call get_cert,./tls/redis,~/certs/production/wikanproductions/redis/${PQSQL_ID})



# ================================  ================================

.PHONY: ca
ca:
	@openssl ecparam -name secp384r1 -genkey -noout -out ${CLEAN_PATH}/service/ca/ca-key.pem > /dev/null 2>&1
	@openssl req -x509 -new -key ${CLEAN_PATH}/service/ca/ca-key.pem -out ${CLEAN_PATH}/service/ca/ca-cert.pem -days 1100 -config ${CLEAN_PATH}/service/ca/ca.cnf > /dev/null 2>&1
	@cat ${CLEAN_PATH}/service/ca/ca-cert.pem ${CLEAN_PATH}/service/ca/ca-key.pem > ${CLEAN_PATH}/service/ca/ca.pem
	@echo "Successfully created CA cert..."

.PHONY: server-cert
server-cert:
	@openssl ecparam -name secp384r1 -genkey -noout -out ${CLEAN_PATH}/service/server/server-key.pem > /dev/null 2>&1
	@openssl req -new -key ${CLEAN_PATH}/service/server/server-key.pem -out ${CLEAN_PATH}/service/server/server-csr.pem -config ${CLEAN_PATH}/service/server/server.cnf > /dev/null 2>&1
	@openssl x509 -req -in ${CLEAN_PATH}/service/server/server-csr.pem -out ${CLEAN_PATH}/service/server/server-cert.pem -CA ${CLEAN_PATH}/service/ca/ca-cert.pem -CAkey ${CLEAN_PATH}/service/ca/ca-key.pem -CAcreateserial -days 1100 -sha384 -extensions v3_req -extfile ${CLEAN_PATH}/service/server/server.cnf > /dev/null 2>&1
	@echo "Successfully creating server certs..."

.PHONY: client-cert
client-cert:
	@openssl ecparam -name secp384r1 -genkey -noout -out ${CLEAN_PATH}/service/client/client-key.pem > /dev/null 2>&1
	@openssl req -new -key ${CLEAN_PATH}/service/client/client-key.pem -out ${CLEAN_PATH}/service/client/client-csr.pem -config ${CLEAN_PATH}/service/client/client.cnf > /dev/null 2>&1
	@openssl x509 -req -in ${CLEAN_PATH}/service/client/client-csr.pem -out ${CLEAN_PATH}/service/client/client-cert.pem -CA ${CLEAN_PATH}/service/ca/ca-cert.pem -CAkey ${CLEAN_PATH}/service/ca/ca-key.pem -CAcreateserial -days 1100 -sha384 -extensions v3_req -extfile ${CLEAN_PATH}/service/client/client.cnf > /dev/null 2>&1
	@echo "Successfully creating client certs..."

.PHONY: clean
clean:
	@find ${CLEAN_PATH} -type f -not \( -name 'client.cnf' -o -name 'server.cnf' -o -name 'ca.cnf' \) -exec rm -f {} +
	@echo "${CLEAN_PATH} path successfully cleaned..."

.PHONY: cert
cert: clean ca server-cert client-cert