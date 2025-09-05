.PHONY: error-pb
# generate error proto
error-pb:
	protoc --proto_path=./errorx --go_out=paths=source_relative:./errorx errorx.proto

.PHONY: optx-pb
# generate optx proto
optx-pb:
	protoc --proto_path=./optx --go_out=paths=source_relative:./optx optx.proto


.PHONY: pagination-pb
# generate pagination proto
pagination-pb:
	protoc --proto_path=./dbx/pagination/ --go_out=paths=source_relative:./dbx/pagination/ pagination.proto
