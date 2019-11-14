package main

const adminindex = `<!DOCTYPE html>
<html>

<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
  <title>Pikari - Admin</title>
  <style>
    * {
      box-sizing: border-box;
      text-align: center;
    }

    table {
      border-collapse: collapse;
      border-spacing: 0px;
      margin-bottom: 1em;
    }

    table,
    th,
    td {
      padding: 5px;
      border: 1px solid black;
    }

    table {
      margin-left: auto;
      margin-right: auto;
    }

    label {
      text-align: right;
    }

    input {
      width: 10em;
    }

    form>input {
      width: 100%;
      max-width: 30em;
      border-width: 2px;
      margin: 3px;
    }

    input::-webkit-outer-spin-button,
    input::-webkit-inner-spin-button {
      -webkit-appearance: none;
      margin: 0;
    }

    input[type=number] {
      -moz-appearance: textfield;
    }

    .dir {
      font-weight: bold;
    }

    .check {
      width: 2em;
    }

    .narrow {
      width: 6em;
    }

    tbody>tr:hover {
      background-color: #ffff99;
    }
  </style>
</head>

<body id="body"></body>
<script src="/pikari.js"></script>
<script>
  const password = window.prompt("Enter admin password")
  if (!password) document.body.innerHTML = "FAIL!"
  else {
    (async function () {
      let AdminModule = await import('/admin/admin.mjs')
      window.admin = new AdminModule.Admin()
      Pikari.start(null, password)
    }())
  }
</script>

</html>`
