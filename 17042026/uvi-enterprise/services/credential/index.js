
const express = require('express');
const app = express();

app.post('/vc',(req,res)=>{
  res.json({vc:"issued", id:"vc123"});
});

app.listen(4200);
