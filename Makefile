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