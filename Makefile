
FLAG 				:=
ifeq ($(OS),Windows_NT)
	FLAG += set
	CLEAN = rmdir /S /q .\out
else
	CLEAN = rm -rf out
endif

## Build:
all: client server ## Build all executables

client: ## Build windows and linux client executables
	$(FLAG) `GOARCH=amd64` && $(FLAG) `GOOS=windows` && go build -o out/bin/winspect.exe ./cmd/winspect/
	$(FLAG) `GOARCH=amd64` && $(FLAG) `GOOS=linux` && go build -o out/bin/winspect ./cmd/winspect/

server: ## Build server executable
	$(FLAG) `GOARCH=amd64` && $(FLAG) `GOOS=windows` && go build -o out/bin/winspectserv.exe ./cmd/winspectserv/

clean: ## Remove build related files
	$(CLEAN)
