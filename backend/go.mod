module github.com/jlewi/p22h/backend

go 1.16

require (
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/language v1.2.0
	cloud.google.com/go/secretmanager v1.0.0 // indirect
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.2
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kubeflow/internal-acls/google_groups v0.0.0-20211220174139-11405888dbb5
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	go.uber.org/zap v1.19.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/api v0.70.0
	google.golang.org/genproto v0.0.0-20220222213610-43724f9ea8cf
	google.golang.org/grpc v1.44.0
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.5
)
