module github.com/ravi-poc/event-store

go 1.23

require (
	github.com/gorilla/websocket v1.5.3
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/ravi-poc/contracts v0.0.0
)

replace github.com/ravi-poc/contracts => ../../contracts
