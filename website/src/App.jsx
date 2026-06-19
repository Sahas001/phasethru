import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Landing from './Landing';
import Auth from './Auth';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/auth" element={<Auth />} />
    </Routes>
  );
}

export default App;
