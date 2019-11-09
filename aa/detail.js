export class Detail {
  constructor(master) {
    this.master = master
    
    e("detailinput").addEventListener('keyup', async function(event) {
      if (event.key != "Enter") return
      if (!master.selectedkey) return alert("Add master name first")      
      if (!await Pikari.setLocks(master.selectedkey)) return
      Pikari.data.get(master.selectedkey).push({"name": e("detailinput").value, "amount": parseInt(e("detailamountinput").value)})
      Pikari.commit()
    })
  }

  render() { document.getElementById("detailol").innerHTML = this.getList() }

  getList() {
    if (!this.master.selectedkey) return ``
    return Pikari.data.get(this.master.selectedkey).reduce((result, item) => result + `
<li id="${item.name}">
  <p>${item.name} ${item.amount}</p>
</li>`, '')
  }
}