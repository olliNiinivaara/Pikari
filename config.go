package main

const tomlconfig = `port = 8080 # IP port
maxpagecount = 10000 # https://sqlite.org/pragma.html#pragma_max_page_count
autodrop = true # if a database update fails (database is full etc): automatically start afresh (true) or exit Pikari (false)
`
