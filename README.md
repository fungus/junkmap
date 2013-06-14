junkmap
=======
Small TCP server for use with a Postfix tcp map.  Allows dynamic creation of throw-away Email addresses that auto-expire after a set period of time.  Also supports ability to manually maintain a list of permanent addresses. 

Utilizes an SQLite3 database to store addresses and timestamps.
go-sqlite3 (https://github.com/mattn/go-sqlite3)

Setup
-----
* Create a JSON configuration file with at least an AddrGood option.  Sample config available at junkmap.json.sample
** Service - Specifies the address and port to listen on.  Default: localhost:2000
** Database - File path to SQLite database.  Default: ./junkmap.db
** ValidTime - Expiry time in Hours.  Default 336 (2 weeks)
** AddrGood - Action and address for a good address.  
   Actions: 200 (success, forward email), 
            400 (temporary fail, try again later), 
            500 (permanent fail, go away)
   Examples: "200 example@example.org" - Forward email to example@example.org
             "500 Go away, I don't like spam" - Tell sender user doesn't exist
** AddrBad - Action and address/message for an expired address (See AddrGood for syntax)

* Configure to run at boot.  Example Upstart config at junkmap.upstart
  The -config option is optional if your config file is call junkmap.json and is in the current dir.
* Configure postfix to read the TCP map.
  Example: virtual_alias_maps = tcp:localhost:2000



