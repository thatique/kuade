NPM_BIN := $(shell readlink -f "web/node_modules/.bin")

.PHONY: build-assets build clean lint test proto

build-assets:
	$(NPM_BIN)/node-sass web/static/sass/application.scss assets/static/css/application.css \
		&& $(NPM_BIN)/postcss assets/static/css/application.css -o assets/static/css/application.css \
			--config web

build: build-assets
	go-bindata -o assets/assets.go -pkg assets assets/...
	go build

clean:
	@rm assets/assets.go
	@rm -rf assets/static

lint:
	go fmt ./...

test:
	go test ./...

PROTOC := protoc
PROTO_INCLUDES := \
	-I proto \
	-I $(GOPATH)/src \
	-I $(GOPATH)/src/github.com/gogo/protobuf/protobuf \
	-I $(GOPATH)/src/github.com/gogo/protobuf

PROTO_GOGO_MAPPINGS := $(shell echo \
		Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/protobuf/google/protobuf, \
		Mkuade.protobuf=github.com/thatique/kuade/proto \
	| sed 's/ //g')

proto:
	$(PROTOC) \
		$(PROTO_INCLUDES) \
		--gogo_out=plugins=grpc,$(PROTO_GOGO_MAPPINGS):$(PWD)/proto/ \
		./proto/model.proto
