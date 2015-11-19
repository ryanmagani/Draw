all: compile

compile:
	go build server.go

clean:
	rm server