
const express = require('express');
const app = express();
app.use(express.json());

app.get('/', (req,res)=>res.send("UVI Gateway Running"));

app.listen(3000, ()=>console.log("Gateway on 3000"));
