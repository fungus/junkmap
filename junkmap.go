package main

import (
    "os"
    "fmt"
    "net"
    "log"
    "flag"
    "time"
    "regexp"
    "strings"
    "encoding/json"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    )

/////////////////////////////////
// Configuration 
/////////////////////////////////
type Settings struct {
    Service string    // Address to listen to
    Database string   // Database file
    ValidTime float64 // Valid lifetime to a record (in hours)
    AddrGood string   // Return for a good address (200 success, 400 temp fail, 500 perm fail)
    AddrBad string    // Return for expired address
}
// Defaults
var CFG = Settings{
    "127.0.0.1:2000",
    "junkmap.db",
    336,
    "200 root",
    "500 Unknown User",
}

func read_conf(confFile string) {
    f, err := os.Open(confFile)
    if err != nil {
        log.Fatalf("Unable to open configuration file %s: %s",confFile,err)
    }
    defer f.Close()
    decoder := json.NewDecoder(f)
    err = decoder.Decode(&CFG)
    if err != nil {
        log.Fatalf("Unable to parse configuration file %s: %s",confFile,err)
    }
    validAddr := regexp.MustCompile(`^[245]00 \w`)
    if !validAddr.MatchString(CFG.AddrGood) {
        log.Fatal("AddrGood config value is not valid: ",CFG.AddrGood)
    }
    if !validAddr.MatchString(CFG.AddrBad) {
        log.Fatal("AddrBad config value is not valid: ",CFG.AddrBad)
    }
}

/////////////////////////////////
// Main
/////////////////////////////////
func main() {
    var cfg string
    flag.StringVar(&cfg,"config","junkmap.json","Specify path to config file")
    flag.Parse()
    // Read config file
    read_conf(cfg)
    // Verify DB exists w/ correct schema
    db_check()
    // Start TCP Listener
    listener, err := net.Listen("tcp", CFG.Service)
    if err != nil {
        log.Fatalf("Cannot listen on %s: %s", CFG.Service, err)
    }

    log.Println("Listening on " + CFG.Service)
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    var buf [4096]byte
    defer conn.Close()
    for {
        n, err := conn.Read(buf[0:])
        if err != nil {
            return
        }
        s := string(buf[:n])
        log.Print(s)
        sa := strings.SplitN(s," ",3)
        switch sa[0] {
        case "get":
            value := lookup(strings.TrimSpace(sa[1]))
            reply(conn, value)
        case "put":
            reply(conn,"500 Not implemented")
        default:
            reply(conn,"500 Invalid Input")
        }
    }
}

// Database lookup
func lookup(key string) string {
    db, err := sql.Open("sqlite3", CFG.Database)
    if err != nil {
        return db_error(err)
    }
    defer db.Close()

    // Get DB record
    var t time.Time
    var p int
    row := db.QueryRow("SELECT time, perm FROM junk WHERE user = ?", key)
    err = row.Scan(&t,&p)
    // New address found, create record
    if err == sql.ErrNoRows {
        insert, _ := db.Prepare("INSERT INTO junk (user) VALUES (?)")
        _, err = insert.Exec(key)
        t = time.Now()
    }
    if err != nil {
        return db_error(err)
    }
    now := time.Now()
    // Fail for non-permanent and expired records
    if p == 0 && now.Sub(t).Hours() > CFG.ValidTime {
        return CFG.AddrBad
    }
    return CFG.AddrGood
}

// Database error, return temporary failure
func db_error(e error) string {
    log.Printf("DB error: %s", e)
    return "400 DB lookup error"
}

func reply(conn net.Conn, s string) {
    s += "\n"
    conn.Write([]byte(s))
    log.Print(s)
}

//////////////////////////////////////////
// Database version checking and creation
//////////////////////////////////////////
const DB_VERSION = 1
const DB_SCHEMA = `
CREATE TABLE junk (
    user VARCHAR(32) PRIMARY KEY, 
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    perm TINYINT(1) NOT NULL DEFAULT 0);
    `
// Check database schema version
func db_check() {
    db, err := sql.Open("sqlite3", CFG.Database)
    if err != nil { 
        log.Fatalf("Unable to open or create datatbase %s: %s", CFG.Database,err)
    }
    var user_version int
    row := db.QueryRow("PRAGMA user_version")
    err = row.Scan(&user_version)
    if err != nil {
        log.Fatalf("Unsupported SQLite version: %s",err)
    }
    db.Close()
    if user_version == 0 {
        log.Println("First use. Creating new database: " + CFG.Database)
        db_create(CFG.Database)
        return
    }
    if user_version < DB_VERSION {
        newdb := CFG.Database + ".new"
        log.Println("New database version required.  Creating a new database at: " + newdb)
        db_create(newdb)
        log.Fatal("Migrate your data to new database, then replace old db w/ new")
    }
}

// Create new database schema
func db_create(dbFile string) {
    db, err := sql.Open("sqlite3",dbFile)
    if err != nil {
        log.Fatalf("Unable to create new database schema: %s",err)
    }
    defer db.Close()
    // Drop existing table
    _, err = db.Exec("DROP TABLE IF EXISTS junk")
    if err != nil {
        log.Fatalf("Unable to create new database schema: %s",err)
    }
    // Create schema
    _, err = db.Exec(DB_SCHEMA)
    if err != nil {
        log.Fatalf("Unable to create new database schema: %s",err)
    }
    // Set database schema version
    _, err = db.Exec(fmt.Sprintf("PRAGMA user_version = %d",DB_VERSION))
    if err != nil {
        log.Fatalf("Unsupported SQLite version: %s",err)
    }
}


