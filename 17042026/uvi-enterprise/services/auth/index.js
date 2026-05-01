
const express = require('express');
const jwt = require('jsonwebtoken');
const app = express();
app.use(express.json());

app.post('/login',(req,res)=>{
  const token = jwt.sign({role:"admin"}, process.env.JWT_SECRET);
  res.json({token});
});

app.get('/secure',(req,res)=>{
  res.json({status:"secure endpoint"});
});

app.listen(4000);
