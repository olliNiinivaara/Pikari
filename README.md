# 🏆 Pikari
- Rapidly build working prototypes - best way to elicitate UX and functional requirements
- Backend-as-a-Service (BaaS) - concentrate on front-end, data is automatically saved and synced
- Simple Javascript API and lots of pragmatic examples - get productive in record time
- Framework agnostic - use your favorite tools or stick to vanillaJs
- Developer-friendly - manage, configure and update everything remotely via web admin GUI
- Learning web programming? - Pikari is a great environment for that, too

## Quick start

1. Create a 2-level deep directory hierarchy for Pikari: Because Pikari *data* directory will be automatically created to the *parent* directory of the executable's directory, put the executable to a subdirectory.
1. Put latest release to the subdirectory (*curl -J -LO `<URL>`* is our friend):
   - Linux/x86-64: <https://github.com/olliNiinivaara/Pikari/releases/download/v0.9-beta/linux-pikari-v09-beta>
   - Windows/x86-64: not yet released
   - MacOS/x86-64: <https://github.com/olliNiinivaara/Pikari/releases/download/v0.9-beta/macos-pikari-v09-beta>
1. For convenience rename the program to pikari (.exe in Windows but [remove .dms by Safari](https://forums.macrumors.com/threads/safari-erroneously-adding-dms-extension-to-downloads.2080108/))
1. In Linux and MacOS, give executable permissions with *chmod u+x* pikari
1. Learn to use Pikari server with [Pikari tutorial](http://github.com/olliNiinivaara/Hellopikari)
1. Learn to write Pikari applications with [API specification](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html)
1. Upload some more [examples](#examples) to Pikari and study them
1. Start writing your own prototypes!

## <a name="examples"></a>Examples

* [Hellopikari](http://github.com/olliNiinivaara/Hellopikari/): Pikari server tutorial with a very simple application
* [Text-Pikari](http://github.com/olliNiinivaara/Text-Pikari/): Synced editing of private and shared data
* [List-Pikari](http://github.com/olliNiinivaara/List-Pikari/): Master-detail tables of data, CRUD and calculated fields
* [TodoMVC-Pikari](http://github.com/olliNiinivaara/TodoMVC-Pikari/): The classic
* [Chat-Pikari](http://github.com/olliNiinivaara/Chat-Pikari/): Messaging between users without a database

## Building from source

1. You need [go](https://golang.org/)
2. You need [git](https://www.git-scm.com/)
3. You need [gcc](https://gcc.gnu.org/)
4. ```go get github.com/olliNiinivaara/Pikari```

## Contribute

Feature & pull requests welcome, this is an open source community effort