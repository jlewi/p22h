// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'document.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DocumentLink _$DocumentLinkFromJson(Map<String, dynamic> json) => DocumentLink(
      doc: json['doc'] as String,
      name: json['name'] as String?,
      description: json['description'] as String?,
      issue_url: json['issue_url'] as String?,
      comment_url: json['comment_url'] as String?,
    );

Map<String, dynamic> _$DocumentLinkToJson(DocumentLink instance) =>
    <String, dynamic>{
      'doc': instance.doc,
      'name': instance.name,
      'issue_url': instance.issue_url,
      'comment_url': instance.comment_url,
      'description': instance.description,
    };
