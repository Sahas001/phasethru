import React, { useState } from 'react';
import Hero from './components/Hero';
import HowItWorks from './components/HowItWorks';
import ProjectDetails from './components/ProjectDetails';
import GetStarted from './components/GetStarted';
import CostModel from './components/CostModel';
import './App.css';

import logoImg from '/phasethru_logo.png';

import { Link } from 'react-router-dom';
import ThemeToggle from './components/ThemeToggle';

function Landing() {
  const [navOpen, setNavOpen] = useState(false);

  return (
    <>
      <nav className="navbar">
        <Link to="/" className="logo">
          <img src={logoImg} alt="PhaseThru" className="logo-img" />
          <span className="logo-text">
            <span className="logo-phase">phase</span>
            <span className="logo-thru">thru</span>
          </span>
        </Link>

        <button
          className="mobile-toggle"
          onClick={() => setNavOpen(!navOpen)}
          aria-label="Toggle navigation"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </button>

        <div className={`nav-links${navOpen ? ' open' : ''}`}>
          <a href="#how-it-works" onClick={() => setNavOpen(false)}>How it works</a>
          <a href="#features" onClick={() => setNavOpen(false)}>Features</a>
          <a href="#pricing" onClick={() => setNavOpen(false)}>Pricing</a>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <ThemeToggle />
            <Link to="/auth" className="btn-fill nav-cta" onClick={() => setNavOpen(false)}>
              Sign In
            </Link>
          </div>
        </div>
      </nav>

      <main>
        <Hero />
        <div className="section-sep" />
        <HowItWorks />
        <ProjectDetails />
        <div className="section-sep" />
        <GetStarted />
        <CostModel />
      </main>

      <footer className="site-footer">
        <div className="footer-inner">
          <div className="footer-left">
            <img src={logoImg} alt="" className="footer-logo" />
            <span className="footer-copy">&copy; 2026 PhaseThru</span>
          </div>
          <div className="footer-links">
            <a href="mailto:contact@phasethru.dev">Contact</a>
          </div>
        </div>
      </footer>
    </>
  );
}

export default Landing;
