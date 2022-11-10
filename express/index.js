const app = require('express')()
var xmlp = require('express-xml-bodyparser');
app.use(xmlp())

let reqs = 0
app.post('/', (req, res) => {
  reqs++ // just a dumb tally to see client side
  let s = JSON.stringify(req.body)
  res.send(`req#: ${reqs} -> ${s}`)
})

app.listen(3030, () => {
  console.log(`Example app listening on port 3030`)
})
