
const express = require('express');
const app = express();

app.post('/score',(req,res)=>{
  res.json({fraudScore: Math.random()});
});

app.listen(4300);
