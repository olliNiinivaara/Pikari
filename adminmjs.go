package main

const adminmjs = `
export class Admin {
  constructor() {
    window.e = (id) => { return document.getElementById(id) }
    window.n = (name) => { return document.getElementsByName(name)[0] }
    Pikari.addChangeListener(() => { this.render() })
  }

  render() {
    const html1 = /*html*/
      ¤<h2><a href="https://github.com/olliNiinivaara/Pikari/">Pikari</a> Administration</h2>
 <table>
   <thead>
     <tr>
       <th class="dir" colspan="2">Upload new application from...</th>
     </tr>
     <tr>
       <td><button onclick="admin.renderDirform()">Local directory</button></td>
       <td><button onclick="admin.renderGitform()">Git repository</button></td>
     </tr>
   </thead>
 </table>
 <table>
   <thead>
     <tr>
       <th>Directory</th>
       <th>Edit...</th>
       <th>Name</th>
       <th>Password</th>
       <th>Maxpagecount</th>
       <th>Autorestart</th>
       <th>Source</th>
     </tr>
   </thead>
   <tbody>¤
    const html2 = Pikari.getFields().sort((a,b) => admin.sort(a, b)).filter((key) => key != "admin").reduce((result, key) => {
      const disabled = Pikari.data.get(key).Disabled == 1 ? "text-decoration: line-through; color:gray;" : ""
      const sel_1 = Pikari.data.get(key).Autorestart == -1 ? "selected" : ""
      const sel1 = Pikari.data.get(key).Autorestart == 1 ? "selected" : ""
      const sel0 = Pikari.data.get(key).Autorestart == 0 ? "selected" : ""
      const maxpc = Pikari.data.get(key).Maxpagecount == -1 ? "" : Pikari.data.get(key).Maxpagecount
      result += /*html*/
    ¤   <tr>
   <td class="dir" style="${disabled}">${key}</td>
   <td><button onclick="admin.renderUpdateform('${key}')">✎</button></td>
   <td><input id="Name${key}" type="text" value="${Pikari.data.get(key).Name}" onchange="admin.change('${key}', 'Name')"></td>
   <td><input id="Password${key}" class="narrow" type="text" value="${Pikari.data.get(key).Password}" onchange="admin.change('${key}', 'Password')"></td>
   <td><input id="Maxpagecount${key}" class="narrow" type="number" min="0" placeholder="default" value="${maxpc}" onchange="admin.change('${key}', 'Maxpagecount')"></td>
   <td><select id="Autorestart${key}" onchange="admin.change('${key}', 'Autorestart')">
       <option value="-1" ${sel_1}>default</option>
       <option value="1" ${sel1}>True</option>
       <option value="0" ${sel0}>False</option>
     </select></td>
   <td>${Pikari.data.get(key).Source}</td>
 </tr>¤
     return result}, ¤¤)
    
    const html3 = /*html*/
¤</tbody>
 </table>
 <table>
   <thead>
     <tr>
       <td>
         <details>
           <summary>Default maxpagecount</summary>
           Maxpagecount sets maximum database size for an application. The quantity is in <a
             href="https://sqlite.org/pragma.html#pragma_max_page_count">sqlite pages</a>.<br>
           Assuming sqlite default page size of 4096 bytes, value 10000 for maxpagecount sets the maximum database size
           limit to around 40 megabytes.<br>
           If an application does not use a database, maxpagecount should be set to 0.<br>
         </details>
       </td>¤
    const html4 = ¤<td><input id="Maxpagecount" class="narrow" type="number" min="0" value="${Pikari.data.get('admin').Maxpagecount}" onchange="admin.changeDefault('Maxpagecount')"></td>¤
    const html5 = /*html*/
      ¤    </tr>
     <tr>
       <td>
         <details>
           <summary style="text-align: start;">Default autorestart</summary>
           If a database update fails (usually because database maxpagecount limit is reached) autorestart will
           automatically drop database and start afresh.<br>
           If autorestart is not set, a failing update will terminate the whole Pikari Server.<br>
         </details>
       </td>¤
    const checked = Pikari.data.get("admin").Autorestart == 1 ? "checked" : ""
    const html6 = ¤<td><input id="Autorestart" class="check" type="checkbox" ${checked} onchange="admin.changeDefault('Autorestart')"></td>¤
    const html7 = /*html*/
      ¤     </tr>
   </thead>
 </table>
 <a href="/?user=${Pikari.user}">Back to index</a>
 ¤
   e("body").innerHTML = html1+html2+html3+html4+html5+html6+html7
  }

  async change(key, prop) {
    const el = e(prop+key)
    if (prop=="Name") {      
      if (admin.nameExists(el.value)) {
        el.value = Pikari.data.get(key)["Name"]
        return alert("Name is already in use")
      }
      if (!el.value) el.value = key
    }
    if (prop=="Maxpagecount" && !e("Maxpagecount"+key).value) el.value = -1
    if (await Pikari.setLocks(key)) {
      Pikari.data.get(key)[prop] = el.type == "number" || el.nodeName == "SELECT" ? parseInt(el.value) : el.value
      Pikari.commit()
    }
    if (e("Maxpagecount"+key).value == -1) e("Maxpagecount"+key).value = ""
  }

  async changeDefault(prop) {
    if (prop=="Autorestart") e("Autorestart").value = e("Autorestart").checked ? 1 : 0
    if (!e(prop).value) e.prop.value = "0"
    if (await Pikari.setLocks("admin")) {
      Pikari.data.get("admin")[prop] = e(prop).valueAsNumber
      Pikari.commit()
    }
  }

  nameExists(name) {
    for (let a of Pikari.data.values()) if (a.Name == name) return true
    return false
  }

  renderDirform() {
    e("body").innerHTML = /*html*/
      ¤<h2>Upload new application from local directory</h2>
  <form id="dirform" style="display: flex; flex-direction: column; align-items: center">
    <input name="dir" placeholder="Server directory to upload to" maxlength="1000"> 
    <input name ="files" type="file" webkitdirectory mozdirectory directory multiple/>
    <input name="source" placeholder="Source (optional reminder)" maxlength="1000">    
    <div><button type="button" onclick="admin.submitDirform()">Proceed</button><button type="button" onclick="admin.render()">Cancel</button></div>
  </form>¤
  }

  async submitDirform() {
    if (!n("dir").value) return alert("Directory missing")
    if (n("files").files.length == 0) return alert("Files missing")
    Pikari.waiting(true)
    const response = await fetch("dirupload", {
      method: 'POST',
      body: new FormData(e("dirform"))
    })
    const text = await response.text()
    Pikari.waiting(false)
    if (text) alert(text)
  }

  renderGitform() {
    e("body").innerHTML = /*html*/
      ¤<h2>Upload new application from remote Git repository</h2>
  <form id="gitform" style="display: flex; flex-direction: column; align-items: center">
    <input name="dir" placeholder="Server directory to upload to" maxlength="1000"> 
    <input name="url" type="url" placeholder="URL to upload from" maxlength="1000">    
    <div><button type="button" onclick="admin.submitGitform()">Proceed</button><button type="button" onclick="admin.render()">Cancel</button></div>
  </form>¤
  }

  async submitGitform() {
    if (!n("dir").value) return alert("Directory missing")
    if (!n("url").value) return alert("Url missing")
    Pikari.waiting(true)
    const response = await fetch("gitupload", {
      method: 'POST',
      body: new FormData(e("gitform"))
    })
    const text = await response.text()
    Pikari.waiting(false)
    if (text) alert(text)
  }

  renderUpdateform(dir) {
    const checked = Pikari.data.get(dir).Disabled == 1 ? "checked" : ""
    const fileinput = !Pikari.data.get(dir).Git ? '<input name="files" type="file" webkitdirectory mozdirectory directory multiple/>' : ""
    const sourcehint = fileinput ? "Source (optional reminder)" : "URL of the remote Fit repository"
    const updategit = fileinput ? "" : '<label>Update<input name="dogit" class="check" type="checkbox" value="checked" checked></label>'
    e("body").innerHTML = /*html*/
      ¤<h2>Update ${Pikari.data.get(dir).Name}</h2>
  <form id="updateform" style="display: flex; flex-direction: column; align-items: center">
    <input name="dir" type="hidden" value="${dir}"/>
    ${fileinput}
    <input name="source" placeholder="${sourcehint}" maxlength="1000" value="${Pikari.data.get(dir).Source}">
    <div style="display: grid">
      ${updategit}
      <label>Disabled<input name="disabled" class="check" type="checkbox" value="checked" ${checked}></label>
      <label>Delete existing data</span><input name="deletedata" class="check" value="checked" type="checkbox"></label>
      <label>Delete application</span><input name="deleteapp" class="check" value="checked" type="checkbox"></label>
    </div>
    <div><button type="button" onclick="admin.submitUpdateform('${dir}')">Proceed</button><button type="button" onclick="admin.render()">Cancel</button></div>
  </form>¤
  }

  async submitUpdateform(dir) {
    let response
    if (n("deleteapp").checked) {
      if (!confirm("Really delete " + Pikari.data.get(dir).Name+ "?")) return
      Pikari.waiting(true)
      response = await fetch("delete?app="+dir)
    } else {
      Pikari.waiting(true)
      response = await fetch("update", {
        method: 'POST',
        body: new FormData(e("updateform"))
      })
    }
    const text = await response.text()
    Pikari.waiting(false)
    if (text) alert(text); else admin.render
  }

  sort (a, b) {
    const A = a.toUpperCase()
    const B = b.toUpperCase()
    if (A < B) return -1
    if (A > B) return 1
    return 0
  }
}`
