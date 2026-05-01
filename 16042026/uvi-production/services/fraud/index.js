
const express = require('express');
const app = express();

app.get('/check', (req,res)=>{
  res.json({fraud:false});
});

app.listen(6000, ()=>console.log("Fraud service"));
