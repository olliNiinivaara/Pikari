<p style="text-align: center;">PIKARI Version 0.8 | 🏆 Pikari server manual</p>

NAME
====

**pikari** — Rapid web application prototyping framework

SYNOPSIS
========
**pikari** **-appdir** _directorypath_ \[**-password** _password_\]

DESCRIPTION
===========

pikari is a web server and a database for the Pikari rapid web application prototyping framework.
It serves as an end-point to *pikari.js* API, documented [here](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html).
It also serves all static files from a given directory path to IP address 127.0.0.1.

At start, pikari looks for *pikari.toml* and *pikari.js* at the *current working directory* (*cwd*).
If a file is not found, it will recreate it.
This means that you can use different, modified *pikari.toml* or *pikari.js* for different applications by starting them at different *cwd*.

Options
-------

-appdir _directorypath_

Defines the directory to serve. The directory should contain *index.html* file which is served as default.    
The appdir can be either relative to *cwd* or absolute path.

\[-password _password_\]  
Optional password that clients must know. The client can set the password at [Pikari.start()](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html#.start).

Note that the password is not encrypted in flight; therefore you MUST use HTTPS/WSS reverse proxy to keep it secret.
</p>


FILES
=====

pikari

The program itself.

*\[cwd/\]pikari.js*

[The front-end API implementation](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html)

*\[appdir/\]index.html*

The default file to serve for an application (you'll write it!)

*\[appdir/\]data.db*

The sqlite3 database for an application.

*\[appdir/\]pikari.log*

Logs for an application. This is a rolling log with at most 2 backup files and 1 megabyte of log data.

*\[cwd/\]pikari.toml*

The configuration file. The available options are described in it.
   
note: Assuming sqlite default page size of 4096 bytes, the default value of 10000 for maxpagecount sets the maximum database size limit to around 40 megabytes.

note: If your prototype is not using a database, set maxpagecount to 0 get rid of it.

note: If database autorestarts itself, the *drop* message's sender (and therefore changeCallback's changer) will be *server autorestart*.


REVERSE PROXY ENVIRONMENT
=========================

You should use a HTTPS+WSS reverse proxy between pikari and client when deploying a prototype to the open Internet. The web socket upgrade request is sent
to path */ws* which you must hail. The configuration for [NGINX](https://www.nginx.com/) (say) should be something like this:

```nginx
upstream pikariws {
  server 127.0.0.1:8080;
  keepalive 100;
}
server {
  location /pikari/ {
    proxy_pass http://127.0.0.1:8080/;
  }
  location /pikari/ws {
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_http_version 1.1;
    proxy_connect_timeout 8h;
    proxy_send_timeout 8h;
    proxy_read_timeout 8h;
    proxy_pass http://pikariws/ws;
  }
```

NOTES
===========

The normal data modification protocol between client and server is as follows:
1. Client sends a *setlocks* message to server to request mutually exclusive write locks for some data (fields)
2. Server sends a *lock* message to all clients that tells which locks are currently held by which client
3. If client was not able to acquire requested locks, it MUST abort the modification procedure. Otherwise:
4. When all modified data is available, client sends a *commit* message to server which contains the modified data fields
5. New data is written (field-by-field inserts, updates and/or deletions) to disk and all locks held by client are released
6. Server sends a *change* message to all clients which contains the modified data fields
7. Server sends a *lock* message to all clients that tells which locks are currently held by which client

Tip: When you change the application served by pikari, you may need to
[hard reload](https://en.wikipedia.org/wiki/Wikipedia:Bypass_your_cache) the browser page

BUGS
====

See GitHub Issues: <https://github.com/olliNiinivaara/Pikari/issues>

EXAMPLES
========

./pikari -appdir Hellopikari

in linux, serves an application at directory _./_Hellopikari_/_ with *pikari*, *pikari.js* and *pikari.toml* at *cwd*.


C:\pikari -appdir C:\\Hellopikari -password pillowHikari

in windows, starts *C:\\pikari* and serves application at absolute path _C:\\Hellopikari_  with *pikari.js* and *pikari.toml* at *cwd* and with password _pillowHikari_.


../pikari -appdir .

Runs *pikari* from parent directory and serves application in *cwd* with it's *pikari.js* and *pikari.toml*.

AUTHOR
======

Olli Niinivaara <olli.niinivaara@verkkoyhteys.fi>

---

<p style="text-align: center;">0.8 2019-27-09</p>