
const express = require('express');
const app = express();
app.use(express.json());

app.post('/issue', (req,res)=>{
  res.json({status:"credential issued"});
});

app.listen(5000, ()=>console.log("Credential service"));
