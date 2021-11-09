
PROGRAM=varuh

all: program

program: $(wildcard *.go)
	@echo "Building ${PROGRAM}"
	@go build -o ${PROGRAM} *.go

install: 
	@echo -n "Installing ${PROGRAM}"
	@if [ -f ./${PROGRAM} ]; then \
		sudo cp ./${PROGRAM} /usr/local/bin; \
		echo "...done"; \
	fi 

clean:
	@echo "Removing ${PROGRAM}"
	@rm -f ${PROGRAM}
