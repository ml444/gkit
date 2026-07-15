
.PHONY: test-dbx
# run the dbx core unit tests and each adapter module's integration contract
test-dbx:
	go test ./dbx/...
	cd dbx/gorm && go test ./...
	cd dbx/sqlx && go test ./...

.PHONY: error-pb
# generate error proto
error-pb:
	protoc --go_out=paths=source_relative:. ./errorx/errorx.proto

.PHONY: optx-pb
# generate optx proto
optx-pb:
	protoc --go_out=paths=source_relative:. ./optx/optx.proto


.PHONY: pagination-pb
# generate pagination proto
pagination-pb:
	protoc --go_out=paths=source_relative:. ./dbx/pagination/pagination.proto
