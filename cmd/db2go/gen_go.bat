@echo off
go build -ldflags "-s -w"

set OUT_DIR=.
set PACK_NAME=dataobject

go run main.go --url "mysql://root:123456@127.0.0.1:3306/test?charset=utf8" --disable-decimal ^
--out %OUT_DIR% --db "test" --table "users, classes" --suffix do --package %PACK_NAME% --readonly "created_at, updated_at"
echo generate go file ok, formatting...
gofmt -w %OUT_DIR%/%PACK_NAME%
pause