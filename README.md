junkmap
=======
Small TCP server for use with a Postfix tcp map.  Allows dynamic creation of throw-away Email addresses that auto-expire after a set period of time.  Also supports ability to manually maintain a list of permanent addresses. 

Utilizes an SQLite3 database to store addresses and timestamps.
go-sqlite3 (https://github.com/mattn/go-sqlite3)

Options
-------
All options are optional, but likely you'll want to specify a *domain* and *address* at minimum.
* --service     - Define the address and port to listen on.  Default: 127.0.0.1:2000
* --database    - Path to the database file.  Default: junkmap.db
* --log         - Path to logfile.  Default: stdout
* --expires     - Time in hours for an address to exist after automatic creation.  Default: 336 (two weeks)
* --address     - Forwarding email address.  Default: root
* --domain      - Domain containing addresses to forward.  Default: example.org

Setup
-----
* Configure to run at boot.  Example Upstart config at junkmap.upstart
* Configure postfix to read the TCP map.
  Example: virtual_alias_maps = tcp:localhost:2000
           virtual_alias_domains = example.org


