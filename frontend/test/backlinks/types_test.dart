// Import the test package
import 'package:test/test.dart';
import 'package:feed/backlinks/types.dart';
import 'package:http/testing.dart' as hTesting;
import 'dart:io';
import 'dart:convert';

void main() {
  test('test getDocLinks()-basic', () async {
    var l = BackLink(text: "sometext", docId: "gdrive.1234");
    expect(l.getDocLink(), "https://docs.google.com/document/d/1234");
  });
  test('test getDocLinks()-empty', () async {
    var l = BackLink(text: "sometext", docId: "1234");
    expect(l.getDocLink(), "");
  });
}
