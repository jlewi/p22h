module github.com/jlewi/p22h/backend

go 1.16

require (
	cloud.google.com/go/secretmanager v1.0.0 // indirect
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.2
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/kubeflow/internal-acls/google_groups v0.0.0-20211220174139-11405888dbb5
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	go.uber.org/zap v1.19.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/api v0.62.0
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.5
)
