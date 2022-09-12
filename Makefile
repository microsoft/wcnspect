ifeq ($(OS),Windows_NT)
	CLIENT_WINDOWS = set `GOARCH=amd64` && set `GOOS=windows` && go build -o out/bin/wcnspect.exe ./cmd/wcnspect/
	CLIENT_LINUX = set `GOARCH=amd64` && set `GOOS=linux` && go build -o out/bin/wcnspect ./cmd/wcnspect/

	SERVER = set `GOARCH=amd64` && set `GOOS=windows` && go build -o out/bin/wcnspectserv.exe ./cmd/wcnspectserv/

	CLEAN = rmdir /S /q .\out
else
	CLIENT_WINDOWS = GOARCH=amd64 GOOS=windows go build -o out/bin/wcnspect.exe ./cmd/wcnspect/
	CLIENT_LINUX = GOARCH=amd64 GOOS=linux go build -o out/bin/wcnspect ./cmd/wcnspect/

	SERVER = GOARCH=amd64 GOOS=windows go build -o out/bin/wcnspectserv.exe ./cmd/wcnspectserv/

	CLEAN = rm -rf out
endif

## Build:
all: client server ## Build all executables

client: ## Build windows and linux client executables
	$(CLIENT_WINDOWS)
	$(CLIENT_LINUX)

server: ## Build server executable
	$(SERVER)

clean: ## Remove build related files
	$(CLEAN)
