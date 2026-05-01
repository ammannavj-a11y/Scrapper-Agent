
const express = require('express');
const app = express();

app.post('/record',(req,res)=>{
  res.json({tx:"mock-blockchain-tx"});
});

app.listen(4400);
