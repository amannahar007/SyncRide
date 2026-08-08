module github.com/syncride/backend/passenger-service

go 1.25.0

require (
	github.com/syncride/backend/shared v0.0.0
	google.golang.org/grpc v1.83.0
	github.com/google/uuid v1.6.0
	gorm.io/gorm v1.31.2
)

require (
	github.com/syncride/backend/shared v0.0.0
	google.golang.org/grpc v1.83.0
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/syncride/backend/shared => ../shared
