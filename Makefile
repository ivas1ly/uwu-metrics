.DEFAULT_GOAL := build

.PHONY:clean
clean:
	-rm -f ./cmd/agent/agent
	-rm -f ./cmd/server/server

.PHONY:statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

.PHONY:iter1
iter1: statictest build
	metricstest -test.v -test.run=TestIteration1$$ \
				-binary-path=cmd/server/server

.PHONY:iter2
iter2: iter1
	metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
                -source-path=. \
                -agent-binary-path=cmd/agent/agent

.PHONY:iter3
iter3: iter2
	metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
                -source-path=. \
                -agent-binary-path=cmd/agent/agent \
                -binary-path=cmd/server/server

.PHONY:iter4
iter4: iter3
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration4$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter5
iter5: iter4
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration5$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter6
iter6: iter5
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration6$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter7
iter7: iter6
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration7$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter8
iter8: iter7
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration8$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter9
iter9: iter8
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration9$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-file-storage-path=$$TEMP_FILE \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter10
iter10: iter9
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration10[AB]$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter11
iter11: iter10
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration11$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter12
iter12: iter11
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration12$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter13
iter13: iter12
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration13$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
				-server-port=$$SERVER_PORT \
				-source-path=.

.PHONY:iter14
iter14: iter13
	SERVER_PORT=$$(random unused-port) ; \
	ADDRESS="localhost:$${SERVER_PORT}" ; \
	TEMP_FILE=$$(random tempfile) ; \
	metricstest -test.v -test.run=^TestIteration14$$ \
				-agent-binary-path=cmd/agent/agent \
				-binary-path=cmd/server/server \
				-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
				-key="$${TEMP_FILE}" \
				-server-port=$$SERVER_PORT \
				-source-path=. ;\
	go test -v -race ./...

.PHONY:build
build:
	go build -C ./cmd/agent/ -o agent
	go build -C ./cmd/server/ -o server

.PHONY:lint
lint:
	golangci-lint run -v