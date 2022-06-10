// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility that Flutter provides. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:feed/backlinks/backlink.dart';
import 'package:feed/backlinks/backlinks.dart';
import 'package:feed/backlinks/documents_service.dart';
import 'dart:io';
import 'dart:convert';
import 'package:http/testing.dart' as hTesting;
import 'package:http/http.dart' as http;

void main() {
  testWidgets('Test the backlinks widget', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const FakeApp());

    // TODO(jeremy): If we don't call pumpAndSettle we get errors
    // about timers still pending. I'm not sure exactly why this works.
    // I suspect it has to do with the fact that are widget is asynchronously
    // loading data to render and so we need to wait for it to finish.
    // somehow pumpAndSettle causes us to wait.
    // see https://docs.flutter.dev/cookbook/testing/widget/introduction
    await tester.pumpAndSettle();

    // Very crude assertion that the links are rendered correctly.
    // Right now the Link text appears twice in each row.
    expect(find.text("Link1Text"), findsNWidgets(2));
    expect(find.text("Link2Text"), findsNWidgets(2));
  });
}

DocumentsService createFakeDocumentsService() {
  Future<http.Response> _mockRequest(http.Request request) async {
    var testData = [
      BackLink(text: "Link1Text", docId: "doc1"),
      BackLink(text: "Link2Text", docId: "doc2")
    ];

    var payload = jsonEncode(testData);
    return http.Response(payload, 200, headers: {
      HttpHeaders.contentTypeHeader: 'application/json',
    });
  }

  return DocumentsService(hTesting.MockClient(_mockRequest));
}

class FakeApp extends StatelessWidget {
  const FakeApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Demo',
      home: Scaffold(
        appBar: AppBar(
          title: Text("Test backlinks"),
        ),
        body: Center(
          child: Backlinks(docs: createFakeDocumentsService()),
        ),
      ),
    );
  }
}
