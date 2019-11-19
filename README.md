# 🏆 Pikari
- Rapidly build working prototypes - best way to elicitate UX and functional requirements
- Backend-as-a-Service (BaaS) - concentrate on front-end, data is automatically served for you
- Simple Javascript API and lots of pragmatic examples - get productive in record time
- Framework agnostic - use your favorite tools or stick to vanillaJs
- Developer-friendly - manage, configure and update everything remotely via web admin GUI
- Learning web programming? - Pikari is a great environment for that, too

## Quick start

1. Create directory for Pikari. Note that Pikari data directory will be created to the *parent* directory of this directory
1. Get suitable binary distribution of latest release (*curl -LJO `<URL>`* works):
   - Linux/x86-64: <https://github.com/olliNiinivaara/Pikari/releases/download/v0.9-alpha/linux_pikari>
   - Windows/x86-64: not yet released
   - Apple/x86-64: not yet released
1. Rename to pikari, and in Linux+MacOS, give program execution permissions: chmod u+x pikari
1. Learn to use Pikari server with [Pikari tutorial](http://github.com/olliNiinivaara/Hellopikari)
1. Learn to write Pikari applications with [API specification](http://htmlpreview.github.io/?https://github.com/olliNiinivaara/Pikari/blob/master/doc/pikari_API.html)
1. Upload some more [examples](#examples) to Pikari and study them
1. Start writing your own prototypes!

## <a name="examples"></a>Examples

* [Hellopikari](http://github.com/olliNiinivaara/Hellopikari/): Pikari server tutorial with a very simple application
* [Text-Pikari](http://github.com/olliNiinivaara/Text-Pikari/): Synced editing of private and shared data
* [List-Pikari](http://github.com/olliNiinivaara/List-Pikari/): Master-detail tables of data, CRUD and calculated fields
* [TodoMVC-Pikari](http://github.com/olliNiinivaara/TodoMVC-Pikari/): The classic
* [Chat-Pikari](http://github.com/olliNiinivaara/Chat-Pikari/): Messaging between users without a database - live demo available

## Building from source

1. You need [go](https://golang.org/)
2. You need [git](https://www.git-scm.com/)
3. You need [gcc](https://gcc.gnu.org/)
4. ```go get github.com/olliNiinivaara/Pikari```