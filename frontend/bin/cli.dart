import 'package:feed/document.dart';

//import '../lib/document.dart';

import 'dart:async';
import 'dart:io';

// A CLI for exercising the dart code for development.
void main(List<String> arguments) {
  // print("running feed cli.");
  // //DocumentLink d = DocumentLink(uri: "http://somelink.com");
  // //print("Document JSON: " + d.toJson()["uri"]);
  // for (final a in arguments) {
  //   print("Argument: " + a);
  // }

  // Try running async
  fetchDocs();
  print("Main is done");
  // TODO(jeremy): Bigquery exports the data as a list. If we used JSONL
  // we could use darts streaming library and process it line by line
  // File('data/results-20211227-133846.json')
  //     .readAsString()
  //     .then(parseResultList);
}
