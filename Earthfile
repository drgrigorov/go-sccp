VERSION 0.7

FROM docker.io/golang:1.21.1-bullseye
WORKDIR /opt/utils

build:

	COPY errors.go go.mod header.go sccp.go sccp_test.go udt.go /opt/utils/
	COPY examples /opt/utils/examples/
	COPY params /opt/utils/params/
	COPY utils /opt/utils/utils/

    RUN GOOS=linux go mod tidy && go build -o go-sccp
    #SAVE ARTIFACT . AS LOCAL build
    SAVE IMAGE go-sccp:test
