#!/bin/bash

# changes types so models compile
sed -i.bak \
  -e 's/null.String/string/' \
  -e 's/null.Uint64/uint64/' \
  -e 's/null.Uint32/uint32/' \
  -e 's/null.Uint16/uint16/' \
  -e 's/null.Uint8/uint8/' \
  -e 's/null.Uint/uint/' \
  -e 's/null.Int64/int64/' \
  -e 's/null.Int32/int32/' \
  -e 's/null.Int16/int16/' \
  -e 's/null.Int8/int8/' \
  -e 's/null.Int/int/' \
  -e 's/null.Time/time.Time/' \
  -e 's/gorm.DeletedAt/time.Time/' \
  -e 's/null.Float32/float32/' \
  -e 's/null.Float64/float64/' ./internal/models/*

# generate with the same swag version pinned in go.mod so docs/docs.go stays
# compatible with the runtime library used by `go build`.
go run github.com/swaggo/swag/cmd/swag@v1.8.6 init -g internal/http/http.go

# revert temp changes
find ./internal/models/ -name "*.bak" -exec sh -c 'mv -f $0 ${0%.bak}' {} \;
