## Build:
client: ## Build client executable
	go build -o out/bin/winspect.exe ./cmd/winspect/

server: ## Build server executable
	go build -o out/bin/winspectserv.exe ./cmd/winspectserv/

all: client server ## Build all executables

clean: ## Remove build related files
	rmdir /S /q .\out