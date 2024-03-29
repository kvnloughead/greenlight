include .envrc

# ============================================================
# HELPERS
# ============================================================

## help: print this help message
.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ============================================================
# DEVELOPMENT
# ============================================================

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/create name=$1: generate new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all 'up' migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running all up migrations'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# ============================================================
# QUALITY CONTROL
# ============================================================

## audit: tidy dependencies and format, vet code and run all tests.
#  Requires staticcheck: `go install honnef.co/go/tools/cmd/staticcheck@latest`
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
#  This command downloads copies of all dependencies into the vendor directory.
#  It's important to run it frequently, including after adding new dependencies,
#  so we've added it as a prereq of the audit command. 
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ============================================================
# BUILD
# ============================================================

## build/api: build the cmd/api application
#  The -ldflags flag is used to reduce binary size by removing symbol tables and
#  some debugging information.
# 
#  The go build execution targets the user's operating system. The second 
#  execution targets a linux amd64 architecture.
# 
.PHONY: build
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s -w' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api
	
# ============================================================
# PRODUCTION
# ============================================================

production_host_ip = '64.23.233.245'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh greenlight@${production_host_ip}

## production/deploy/api: deploy the api to production server
#  
#  This command 
#
#    - copies the linux/amd64 api binary, and migration files, 
#      and the systemd unit file to the user directory on the 
#      server
#    - runs the migration files agains the psql database
#		 - moves the unit file to /etc/systemd/system
#    - enables and restarts the api service with systemd
#
#  The GREENLIGHT_DB_DSN environmental variable should already
#  be set on the server. 
#
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP --delete ./migrations greenlight@${production_host_ip}:~
	rsync -P ./remote/production/api.service greenlight@${production_host_ip}:~
	ssh -t greenlight@${production_host_ip} '\
		migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
	'

