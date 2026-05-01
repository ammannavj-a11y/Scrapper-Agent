
const express = require('express');
const jwt = require('jsonwebtoken');
const app = express();
app.use(express.json());

app.post('/login', (req,res)=>{
  const token = jwt.sign({user:"demo"}, process.env.JWT_SECRET);
  res.json({token});
});

app.listen(4000, ()=>console.log("Auth service"));
