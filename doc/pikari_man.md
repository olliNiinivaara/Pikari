<p style="text-align: center;">PIKARI Version 0.8 | 🏆 Pikari server manual</p>

NAME
====

**pikari** — Rapid web application prototyping framework

SYNOPSIS
========
**pikari** **-appdir** _directorypath_ \[**-password** _password_\]

DESCRIPTION
===========

Pikari is a web server and database for the Pikari rapid web application prototyping framework.
It serves files of application directory given as command line parameter to 127.0.0.1.
Above all it serves as an end-point to *pikari.js* API, documented [here](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html).

At start, pikari looks for configuration file *pikari.toml* and the front-end API implementation *pikari.js* files at the *current working directory* (*cwd*).
If a file is not found, it will recreate it.
This means that you can use different, modified *pikari.toml* or *pikari.js* for different applications by starting them at different *cwd*.

Options
-------

-appdir _directorypath_

Defines the application directory for the application to run. The directory should contain *index.html* file which is served as default.    
The appdir can be either relative to *cwd* or absolute.

\[-password _password_\]  
Optional parameter that defines a password string that must be present in every call to server. The client can set the password at [pikari.start()](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html#.start).

Note that the password is not encrypted in flight; therefore you MUST use HTTPS/WSS reverse proxy to gain any real protection in the Internet.
</p>


FILES
=====

pikari | pikari.exe

The program itself. In linux it is named *pikari*.

*\[cwd/\]pikari.js*

The front-end API implementation, see [this](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html)

*\[appdir/\]index.html*

The default file to serve for an application (you'll write it!)

*\[appdir/\]data.db*

The sqlite3 database for an application.

*\[appdir/\]pikari.log*

Logs for an application. This is a rolling log with at most 2 backup files and 1 megabyte of log data.

*\[cwd/\]pikari.toml*

The configuration file. The available three options are described in it.
   
note: Assuming sqlite default page size of 4096 bytes, the default value of 10000 for maxpagecount sets the maximum database size limit to around 40 megabytes.

note: If your prototype is not using a database, set maxpagecount to 0 get rid of it.

note: If database autorestarts itself, the *drop* message's sender user name will be *server autorestart*.

NOTES
===========

The normal data modification protocol between client and server is as follows:
1. Client sends a *setlocks* message to server to request exclusive write locks for some data (fields)
2. Server sends a *lock* message to all clients that tells which locks are currently held by which client
3. If client was not able to acquire requested locks, it MUST abort the modification procedure. Otherwise:
4. When all modified data is available, client sends a *commit* message to server which contains the modified data fields
5. New data is written (field-by-field inserts, updates and/or deletions) to disk and all locks held by client are released
6. Server sends a *change* message to all clients which contains the modified data fields
7. Server sends a *lock* message to all clients that tells which locks are currently held by which client

BUGS
====

See GitHub Issues: <https://github.com/olliNiinivaara/Pikari/issues>

EXAMPLES
========

./pikari -appdir Hellopikari

in linux, serves an application at directory _./_Hellopikari_/_ with *pikari*, *pikari.js* and *pikari.toml* at *cwd*.


C:\pikari -appdir C:\\_Hellopikari -password pillowHikari

in windows, starts pikari at *C:\\pikari* and serves application at absolute path _C:\\_Hellopikari_  with *pikari.js* and *pikari.toml* at *cwd* and with password _pillowHikari_.

AUTHOR
======

Olli Niinivaara <olli.niinivaara@verkkoyhteys.fi>

---

<p style="text-align: center;">0.8 2019-24-09</p>