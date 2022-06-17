# Front End

A frontend for the knowledge graph built ontop of flutter.

## Prerequisites

Install flutter on your machine.

## Running the flutter app in dev mode

```
cd frontend
flutter run
```

## CLI

There is a CLI program meant for testing/developing snippets of code. To run
the CLI

```
dart run bin/cli.dart
```

## Run the server locally

To run locally

```
make run
```

This will build and run the server. The server will be serving the frontend at /ui/

When working on the front end you can start a separate copy of the frontend
by run `flutter run` or running it in vscode in debug mode. Just set
the URL of the backend to whatever port the backend is running on (should be 8080)
by default. Right now this requires hacking the code.

## Why use Flutter & Dart rather than Javascript/Typescript/ReactX

I think the main selling point of Flutter is that you can write an app
once and compile it to a webapp or native Android or IOS application.

The main reason we chose Flutter & Dart is the hope that it is actually
more natural for backend engineers to learn.

Some thoughts (which may not prove to be true)

* Flutter doesn't expose developers to CSS or the HTML DOM
  * Hopefully this simplifies how to control the UI

* Flutter does layout differently from html [reference](https://docs.flutter.dev/development/ui/layout/constraints)
  * Not sure yet what to make of this

## Dart/Flutter cheat sheet

To add dependencies

```
dart pub add <dependency>
```
## References

Flutter references
[flutter online documentation](https://flutter.dev/docs)
