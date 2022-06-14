import 'package:json_annotation/json_annotation.dart';

import 'dart:convert';
import 'package:http/http.dart' as http;
part 'types.g.dart';

// Following: https://pub.dev/packages/json_serializable define
// a class to contain the JSON information
//
// To generate the generated code run
// flutter packages pub run  build_runner build
// see: https://github.com/google/json_serializable.dart/tree/master/example
@JsonSerializable()
class BackLink {
  /// The generated code below handles if the corresponding JSON value doesn't
  /// exist or is empty.
  /// name of the google doc
  final String? text;
  final String? docId;

  BackLink({
    this.text,
    this.docId,
  });

  /// Connect the generated [_$BackLinkFromJson] function to the `fromJson`
  /// factory.
  factory BackLink.fromJson(Map<String, dynamic> json) =>
      _$BackLinkFromJson(json);

  /// Connect the generated [_$BackLinkToJson] function to the `toJson` method.
  Map<String, dynamic> toJson() => _$BackLinkToJson(this);

  String getDocLink() {
    // split string
    List<String> pieces = docId!.split(".");
    if (pieces.length <= 1) {
      return "";
    }
    return "https://docs.google.com/document/d/${pieces[1]}";
  }
}

@JsonSerializable()
class BackLinkList {
  /// The generated code below handles if the corresponding JSON value doesn't
  /// exist or is empty.
  /// name of the google doc
  final List<BackLink>? items;

  BackLinkList({
    this.items,
  });

  /// Connect the generated [_$BackLinkListFromJson] function to the `fromJson`
  /// factory.
  factory BackLinkList.fromJson(Map<String, dynamic> json) =>
      _$BackLinkListFromJson(json);

  /// Connect the generated [_$BackLinkListToJson] function to the `toJson` method.
  Map<String, dynamic> toJson() => _$BackLinkListToJson(this);
}
