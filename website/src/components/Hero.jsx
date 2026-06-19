import React, { useEffect, useRef } from 'react';
import './Hero.css';
import logoImg from '/phasethru_logo.png';

const Hero = () => {
  const termRef = useRef(null);

  useEffect(() => {
    const lines = termRef.current?.querySelectorAll('.tl');
    if (!lines) return;
    lines.forEach((line, i) => {
      line.style.animationDelay = `${i * 0.35 + 0.5}s`;
    });
  }, []);

  return (
    <section className="hero">
      {/* Ambient glow behind the logo */}
      <div className="hero-glow" />

      <div className="hero-center">
        <img src={logoImg} alt="PhaseThru" className="hero-logo" />
        <h1>
          Tunnels that just work.
        </h1>
        <p className="hero-sub">
          Expose any local port to a public URL in seconds.
          No config files, no daemons — one Go binary
          that punches through NAT, CGNAT, and firewalls.
        </p>
        <div className="hero-actions">
          <a href="#get-started" className="btn-fill">Get Started</a>
          <a
            href="#how-it-works"
            className="btn-ghost"
          >
            Learn More
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
              <path d="M4 12L12 4M12 4H6M12 4V10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </a>
        </div>
      </div>

      <div className="hero-terminal-area">
        <div className="terminal" ref={termRef}>
          <div className="terminal-chrome">
            <span className="td td-r" />
            <span className="td td-y" />
            <span className="td td-g" />
            <span className="terminal-title">zsh</span>
          </div>
          <div className="terminal-screen">
            <div className="tl"><span className="t-p">$</span> phasethru http 3000</div>
            <div className="tl t-dim">connecting to relay...</div>
            <div className="tl t-green">connected</div>
            <div className="tl t-dim">session usr_k8m2x9p1</div>
            <div className="tl">
              <span className="t-purple">https://myapp.phasethru.dev</span>
              <span className="t-dim"> → localhost:3000</span>
            </div>
            <div className="tl t-dim t-cursor">ready for connections</div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
