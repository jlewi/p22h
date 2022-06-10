import 'package:json_annotation/json_annotation.dart';

import 'dart:convert';
import 'package:http/http.dart' as http;
part 'document.g.dart';

// Following: https://pub.dev/packages/json_serializable define
// a class to contain the JSON information
//
// To generate the generated code run
// flutter packages pub run  build_runner build
// see: https://github.com/google/json_serializable.dart/tree/master/example
@JsonSerializable()
class DocumentLink {
  /// The generated code assumes these values exist in JSON.
  /// doc is the URI of the document
  final String doc;

  /// The generated code below handles if the corresponding JSON value doesn't
  /// exist or is empty.
  /// name of the google doc
  final String? name;

  final String? issue_url;
  final String? comment_url;

  /// description for the google doc
  final String? description;

  DocumentLink(
      {required this.doc,
      this.name,
      this.description,
      this.issue_url,
      this.comment_url});

  /// Connect the generated [_$DocumentLinkFromJson] function to the `fromJson`
  /// factory.
  factory DocumentLink.fromJson(Map<String, dynamic> json) =>
      _$DocumentLinkFromJson(json);

  /// Connect the generated [_$DocumentLinkToJson] function to the `toJson` method.
  Map<String, dynamic> toJson() => _$DocumentLinkToJson(this);
}

// parseResultList takes a string containing a JSON list of items and
// parses it into a list of DocumentLink.
List<DocumentLink> parseResultList(String contents) {
  List<DocumentLink> docs = [];
  List<dynamic> results = [];

  results = jsonDecode(contents);
  for (final r in results) {
    if (r["doc"] == null) {
      continue;
    }
    //var l = jsonEncode(r);
    // print("Parsing $r");
    var d = DocumentLink.fromJson(r);
    docs.add(d);
  }
  print("Length of results: ${docs.length}");
  return docs;
}

Future<List<DocumentLink>> fetchDocs() async {
  final response = await http.get(Uri.parse(
      'https://raw.githubusercontent.com/jlewi/code-intelligence/flutter/feed/data/results-20211227-133846.json'));

  if (response.statusCode == 200) {
    // If the server did return a 200 OK response,
    // then parse the JSON.
    return parseResultList(response.body);
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('Failed to load docs');
  }
}
