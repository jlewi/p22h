// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'types.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BackLink _$BackLinkFromJson(Map<String, dynamic> json) => BackLink(
      text: json['text'] as String?,
      docId: json['docId'] as String?,
    );

Map<String, dynamic> _$BackLinkToJson(BackLink instance) => <String, dynamic>{
      'text': instance.text,
      'docId': instance.docId,
    };

BackLinkList _$BackLinkListFromJson(Map<String, dynamic> json) => BackLinkList(
      items: (json['items'] as List<dynamic>?)
          ?.map((e) => BackLink.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$BackLinkListToJson(BackLinkList instance) =>
    <String, dynamic>{
      'items': instance.items,
    };
