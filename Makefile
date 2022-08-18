run:
	go run cmd/main.go

up:
	migrate -path migrations -database "postgres://bkfhqnkt:tY-yRBFFiy48j622xXrPO28g0OT-xKsM@abul.db.elephantsql.com/bkfhqnkt" -verbose up

down:
	migrate -path migrations -database "postgres://bkfhqnkt:tY-yRBFFiy48j622xXrPO28g0OT-xKsM@abul.db.elephantsql.com/bkfhqnkt" -verbose down
