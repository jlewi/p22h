import 'package:flutter/material.dart';
import 'package:feed/backlinks/backlinks.dart';

class NavDrawer extends StatelessWidget {
  // Devide a side navigation bar
  // Code originally from:
  // https://maffan.medium.com/how-to-create-a-side-menu-in-flutter-a2df7833fdfb
  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: ListView(
        padding: EdgeInsets.zero,
        children: <Widget>[
          DrawerHeader(
            child: Text(
              'Side menu',
              style: TextStyle(color: Colors.white, fontSize: 25),
            ),
            decoration: BoxDecoration(
                color: Colors.green,
                image: DecorationImage(
                    fit: BoxFit.fill,
                    image: AssetImage('assets/images/cover.jpg'))),
          ),
          ListTile(
            leading: Icon(Icons.input),
            title: Text('Welcome'),
            onTap: () => {},
          ),
          ListTile(
            leading: Icon(Icons.verified_user),
            title: Text('Back Links'),
            onTap: () {                
                // This gets rid of the side bar popup window
                Navigator.of(context).pop();                 
                Navigator.pushNamed(
                  context, "/backlinks");
            },
          ),
          ListTile(
            leading: Icon(Icons.settings),
            title: Text('GitHub Feed'),
            onTap: () {
                // This gets rid of the side bar popup window
                Navigator.of(context).pop();                 
                // TODO(jeremy): Add a widget for the feed. 
            },
          ),
        ],
      ),
    );
  }
}