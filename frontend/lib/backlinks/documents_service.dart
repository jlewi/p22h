import 'package:feed/backlinks/types.dart';
import 'package:http/http.dart' as http;
import 'package:uri/uri.dart';
import 'dart:convert';

class DocumentsService {
  // Inject the http client. This allows mocks to be provided in tests.
  final http.Client client;

  DocumentsService(this.client);

  // DocumentsService is a class for interacting with documents
  // in the backend.
  Future<BackLinkList> getBackLinks(String doc) async {
    // TODO(jeremy): How do we configure the address of the backend
    // TODO(jeremy): The document to fetch the backlinks for shouldn't be hardcoded we should get it from the input box?

    // We need to get the URL the frontend is being served on.
    // The assumption is that the same server is also serving the backend.
    // We need to strip out the path e.g "/ui" and then add the URL
    // of the documents method to call.
    // This most likely would not work behind a reverse proxy.
    var base = Uri.base;
    var builder = UriBuilder.fromUri(base);
    builder.path = "documents/" + doc + ":backLinks";
    // This is a hack during debug mode.
    builder.port = 8080;
    builder.fragment = "";
    var u = builder.build();

    print("calling getBacklinks: URI: " + u.toString());

    final response = await client.get(u);

    if (response.statusCode == 200) {
      // If the server did return a 200 OK response,
      // then parse the JSON.
      print("getBackLinks returned 200");
      // First decode it to Map<String, Dynamic>
      var results = jsonDecode(response.body);
      // Now parse the actual class
      return BackLinkList.fromJson(results);
    } else {
      // If the server did not return a 200 OK response,
      // then throw an exception.
      print("Failed to get backlinks");
      throw Exception('Failed to load backLinks');
    }
  }
}
