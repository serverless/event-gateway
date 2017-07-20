build:
	@gox -os="linux darwin windows" -arch="386 amd64"
.PHONY: build
