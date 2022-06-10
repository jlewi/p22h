build-backend:
	mkdir -p build/bin
	cd backend && go build -o ../build/bin/server ./cmd/...


# Regenerate the generated classes for JSON serialization
flutter-json:
	cd frontend && flutter packages pub run  build_runner build --delete-conflicting-outputs

build-frontend:
	# Build a release version
	# set base-href
	cd frontend && flutter build web --base-href=/ui/


# Run a release version of the application served
# by our backend server.
run: build-backend build-frontend
	build/bin/server --debug=true run --static-path ./frontend/build/web