// Import the test package and Counter class
import 'package:test/test.dart';
import 'package:feed/backlinks/backlink.dart';
import 'package:feed/backlinks/documents_service.dart';
import 'package:http/testing.dart' as hTesting;
import 'package:http/http.dart' as http;
import 'dart:io';
import 'dart:convert';

void main() {
  test('test getBacklinks()', () async {
    Future<http.Response> _mockRequest(http.Request request) async {
      var testData = BackLinkList(items: [
        BackLink(text: "sometext", docId: "doc1"),
        BackLink(text: "someother", docId: "doc2")
      ]);

      var payload = jsonEncode(testData);
      return http.Response(payload, 200, headers: {
        HttpHeaders.contentTypeHeader: 'application/json',
      });
    }

    var documents = DocumentsService(hTesting.MockClient(_mockRequest));

    var links = await documents.getBackLinks("gdrive.1zTy");
    var items = links.items!;
    expect(items.length, 2);
  });
}
