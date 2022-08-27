ifeq ($(OS),Windows_NT)
	CLIENT_WINDOWS = set `GOARCH=amd64` && set `GOOS=windows` && go build -o out/bin/winspect.exe ./cmd/winspect/
	CLIENT_LINUX = set `GOARCH=amd64` && set `GOOS=linux` && go build -o out/bin/winspect ./cmd/winspect/

	SERVER = set `GOARCH=amd64` && set `GOOS=windows` && go build -o out/bin/winspectserv.exe ./cmd/winspectserv/

	CLEAN = rmdir /S /q .\out
else
	CLIENT_WINDOWS = GOARCH=amd64 GOOS=windows go build -o out/bin/winspect.exe ./cmd/winspect/
	CLIENT_LINUX = GOARCH=amd64 GOOS=linux go build -o out/bin/winspect ./cmd/winspect/

	SERVER = GOARCH=amd64 GOOS=windows go build -o out/bin/winspectserv.exe ./cmd/winspectserv/

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
