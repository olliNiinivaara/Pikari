package main

const tomlconfig = `port = 8080 # IP port
maxpagecount = 10000 # https://sqlite.org/pragma.html#pragma_max_page_count; 0 == no database
autorestart = true # if a database update fails (database is full etc): automatically drop database and start afresh (true) or exit Pikari (false)
`
