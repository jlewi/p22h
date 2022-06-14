import 'package:feed/backlinks/documents_service.dart';
import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:feed/backlinks/types.dart';
import 'dart:async';

class Backlinks extends StatefulWidget {
  // GetBackLinksFunc getBacklinks;
  final DocumentsService docs;
  late StreamController<String> docStream;
  // This widget is intended to let you lookup
  // a document and find all the links to it.
  // This is a stateful widget owing to the text field
  // Uses formal parameters to initialize docs variable.
  Backlinks({Key? key, required DocumentsService this.docs}) : super(key: key) {
    docStream = StreamController<String>();
  }

  @override
  State<Backlinks> createState() => _BacklinksWidgetState();
}

class _BacklinksWidgetState extends State<Backlinks> {
  late TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    // Initialize the text editing controller and set an initial value.
    // TODO(jeremy): We should come up with a better way of initializing
    // it rather than hardcoding a specific value that only makes sense
    // for me.
    _controller = TextEditingController()
      ..text = "gdrive.1zT1yhDgmS59_uMWJ76hr7FCyXTDQ1Ju1MVADnlCfNEQ";
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    Widget textBox = TextField(
      controller: _controller,
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: "Document to lookup",
      ),
      onSubmitted: (String value) async {
        // Send the newly entered value to the doc stream.
        widget.docStream.add(value);
      },
    );

    Widget mainColumn = Container(
      padding: const EdgeInsets.all(32),
      child: Column(
        children: [
          Text(
              "Enter the Document to lookup; e.g. gdrive.1zT1yhDgmS59_uMWJ76hr7FCyXTDQ1Ju1MVADnlCfNEQ",
              style: TextStyle(
                color: Colors.grey[500],
              )),
          textBox,
          Text("Results:",
              style: TextStyle(
                color: Colors.grey[500],
              )),
          Expanded(
              child: Results(
                  docs: widget.docs, docStream: widget.docStream.stream)),
        ],
      ),
    );

    return Scaffold(
      appBar: AppBar(
        title: Text('Backlinks'),
      ),
      // Create a text box to get the document to get the links for.
      body: Center(
        child: mainColumn,
      ),
    );
  }
}

class Results extends StatefulWidget {
  // N.B. Data that needs to be injected into the state class is stored in
  // the widget. The state class can use its member variable widget to
  // access variables stored in the corresponding StatefulWidget class
  final DocumentsService docs;
  // docStream is a string of which doc to process
  final Stream docStream;
  const Results(
      {Key? key,
      required DocumentsService this.docs,
      required Stream this.docStream})
      : super(key: key);

  // This widget shows the list of backlinks selected. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  @override
  State<Results> createState() => _ResultsState();
}

class _ResultsState extends State<Results> {
  Future<BackLinkList>? futureLinks;

  // setDoc sets the document to fetch the links for
  void setDoc(String doc) {
    setState(() {
      // This call to setState tells the Flutter framework that something has
      // changed in this State, which causes it to rerun the build method below
      // so that the display can reflect the updated values. State is
      // always changed inside a function passed to setState().
      //
      // In this case we want to change Future<BackLinkList> to use
      // whatever doc has been entered into the text baox
      futureLinks = widget.docs.getBackLinks(doc);
    });
  }

  @override
  void initState() {
    super.initState();
    // Setup a subscription which will call setDoc each time a doc is entered
    widget.docStream.listen((docName) {
      setDoc(docName);
    });
  }

  // _buildLinkRow returns a row to render a link.
  Widget _buildLinkRow(BackLink link) {
    String text = "";
    if (link.text != null) {
      text = link.text as String;
    }

    String docId = "";
    if (link.docId != null) {
      docId = link.docId as String;
    }

    // Create a row to display the backlink information.
    Widget row = Container(
      padding: const EdgeInsets.all(32),
      child: Row(
        children: [
          Expanded(
            /* Column 1 is the text of the link*/
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                /*2*/
                Container(
                  padding: const EdgeInsets.only(bottom: 8),
                  // N.B. Could also use RichText
                  // https://stackoverflow.com/questions/43583411/how-to-create-a-hyperlink-in-flutter-widget
                  //
                  // N.B. when you however over it you don't get the link.
                  // You also can't right click and copy the link.
                  child: new InkWell(
                      child: new Text(
                        text,
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      onTap: () => launch(link.getDocLink())),
                ),
                Text(
                  link.docId!,
                  style: TextStyle(
                    color: Colors.grey[500],
                  ),
                ),
              ],
            ),
          ),
          /*3*/
          Icon(
            Icons.document_scanner,
            color: Colors.red[500],
          ),
          _buildLinkWidget(text, link.getDocLink()),
        ],
      ),
    );
    return row;
  }

  // Return a widget with a link
  Widget _buildLinkWidget(text, link) {
    // show link on hover
    Widget row = new InkWell(
      child: new Text(
        text,
        style: TextStyle(
          fontWeight: FontWeight.bold,
        ),
      ),
      onTap: () => launch(link),
      onHover: (event) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(link),
          ),
        );
      },
    );
    return row;
  }

  @override
  Widget build(BuildContext context) {
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.

    if (futureLinks == null) {
      return Text("No results to load. Select a doc to load backlinks for");
    }

    return Scaffold(
      // TODO(jeremy): Should we also be defining the sidebar
      // here? How do we avoid circular dependencies if
      // nav-drawer imports backlink?
      // drawer: NavDrawer(),
      body: Center(
        child: FutureBuilder<BackLinkList>(
          future: futureLinks,
          builder: (context, snapshot) {
            if (snapshot.hasData) {
              List<Widget> rows = [];
              var bList = snapshot.data!;
              for (BackLink r in bList.items!) {
                rows.add(_buildLinkRow(r));
              }

              // Use a list view to handle scrolling.
              // https://api.flutter.dev/flutter/widgets/ListView-class.html
              // This constructor loads all of the items. This is fine
              // for small number of items but for larger items we should
              // probably be loading them on demand.
              ListView view = ListView(
                padding: const EdgeInsets.all(8),
                children: rows,
              );
              return view;
            } else if (snapshot.hasError) {
              return Text('${snapshot.error}');
            }

            // By default, show a loading spinner.
            return const CircularProgressIndicator();
          },
        ),
      ),
    );
  }
}
