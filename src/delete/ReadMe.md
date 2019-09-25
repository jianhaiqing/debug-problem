
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o deletesql.linux deletesql.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o deletesql.mac deletesql.go