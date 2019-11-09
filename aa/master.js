import { Detail } from "./detail.js";

export class Master {
  constructor() {
    const self = this
    this.detail = new Detail(this)    
    this.selectedkey = null

    e("clearbutton").addEventListener('click', function() { if (confirm("Really?")) Pikari.dropData() })

    e("masterinput").addEventListener('change', async function() {
      const name = Pikari.clean(e("masterinput").value)
      if (Pikari.data.has(name)) return alert("Name already exists")
      if (!await Pikari.setLocks(name)) return
      self.selectedkey = name
      Pikari.data.set(name, [])
      e("masterinput").value = ""
      Pikari.commit()
    })

    Pikari.addChangeListener(function() {
      if (self.selectedkey && !Pikari.data.has(self.selectedkey)) self.selectedkey = null
      if (!self.selectedkey && Pikari.data.size > 0) self.selectedkey = Pikari.data.keys().next().value
      self.render()
    })  
  }

  render() {
    e("masterol").innerHTML = this.getList()
    const items = document.getElementsByClassName("masterli")
    const self = this
    for (let elem of items) {
      elem.addEventListener("click", function() { 
        self.selectedkey = Pikari.clean(elem.id)
        self.render()
      })
    }

    if (e("delete")) e("delete").addEventListener("click", async function() {
      if (!await Pikari.setLocks(self.selectedkey)) return
      Pikari.data.set(self.selectedkey, null)
      Pikari.commit()
    })

    this.detail.render(this.selectedkey)
    let sumtotal = 0
    Pikari.data.forEach(value => { for (let detail of value) sumtotal+=detail.amount })
    e("sumtotal").innerHTML = "Sumtotal: "+sumtotal
  }

  getList() {
    const buttondelete = "<button id='delete' title='delete'>x</button>"
    return Pikari.getFields().reduce((result, key) => result + `
<li id ="${key}" class="masterli ${key == this.selectedkey ? " selected" : ""}">
  <span style="display: inline-flex">
  <p>${key}&nbsp;</p>
  <p>${Pikari.data.get(key).reduce((result, d) => result + d.amount, 0)}&nbsp;</p>
  ${key == this.selectedkey ? buttondelete : ""}
  </span>
</li>`, '')
  }
}