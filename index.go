package main

const index1 = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
  <title>Pikari</title>
  <noscript>Pikari needs Javascript in order to work.</noscript>
</head>
<body>
  <h2>Pikari</h2>
  <ol id="applist"></ol>
</body>  
<script src="pikari.js"></script>
<script>
  let username = new URLSearchParams(window.location.search).get('user')  
  if (!username) username = window.prompt("Enter your user name for Pikari")
  if (!username) document.body.innerHTML = "FAIL!"
  else {
    const sort = function(a,b) {
      const A = Pikari.data.get(a).toUpperCase() ; const B = Pikari.data.get(b).toUpperCase()
      if (A < B) return -1 ; if (A > B) return 1 ; return 0
    }
    Pikari.addChangeListener(()=>{
			let url = location.href
      const end = url.indexOf("?")
      if (end > -1) url = url.substring(0, end)	
		  document.getElementById("applist").innerHTML = Pikari.getFields().sort((a,b) => sort(a, b)).reduce((result, key) => result +`

const index2 = `<li><a href="${url+key+'/?user='+Pikari.user}">${Pikari.data.get(key)}</a></li>`

const index3 = `, '')
		})
		Pikari._app = "index" // hack, do not remove
    Pikari.start(username)
  }
</script>
</html>`
