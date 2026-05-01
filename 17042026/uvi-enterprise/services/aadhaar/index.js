
const express = require('express');
const app = express();

app.get('/verify',(req,res)=>{
  res.json({aadhaar:"verified", score:0.98});
});

app.listen(4100);
