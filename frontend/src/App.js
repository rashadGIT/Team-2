import React, {useEffect, useState, useInput }from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import './App.css';

function App() {
  const [isURLVisible, setIsURLVisible] = useState(false);
  
  const generateURL = () => {
    setIsURLVisible(!isURLVisible)
  }

  return (
    <div className="App">
      <header className="App-header">
        <TextField
          required
          id="outlined-required"
          label="Required"
          defaultValue="Enter Name"
        />
        <br />
        <Button variant="contained" size="small" onClick={() => generateURL()}>
          Generate Game URL
        </Button>
        {(isURLVisible) && 
        <p>
          Hello World
        </p>}
      </header>
    </div>
  );
}

export default App;
