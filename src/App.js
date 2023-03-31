import React from 'react';
import './App.css';

import { Amplify } from 'aws-amplify';
import awsconfig from './aws-exports';
import Quote from './components/Quote';

Amplify.configure(awsconfig);


function App() {
  return (
    <Quote/>
  );
}

export default App;
