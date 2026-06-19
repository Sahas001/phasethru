import React, { useState } from 'react';
import './GetStarted.css';

const GetStarted = () => {
  const [copied, setCopied] = useState(false);
  const installCmd = 'curl -fsSL https://phasethru.dev/install.sh | sh';

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <section id="get-started" className="getstarted">
      <div className="wrap">
        <div className="gs-inner">
          <div className="gs-text">
            <div className="gs-label">Get started</div>
            <h2>Up and running<br />in under a minute.</h2>
            <p>
              Install the CLI, authenticate once, and start tunneling.
              Works on macOS, Linux, and WSL.
            </p>
          </div>

          <div className="gs-install">
            <div className="gs-step">
              <span className="gs-step-num">1</span>
              <span className="gs-step-label">Install</span>
            </div>
            <div className="gs-cmd-wrap" onClick={handleCopy} title="Click to copy">
              <code className="gs-cmd">{installCmd}</code>
              <span className="gs-copy-icon">
                {copied ? (
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                    <path d="M3.5 8.5L6.5 11.5L12.5 4.5" stroke="var(--green)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                ) : (
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                    <rect x="5" y="5" width="8" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.2"/>
                    <path d="M11 5V3.5C11 2.67 10.33 2 9.5 2H3.5C2.67 2 2 2.67 2 3.5V9.5C2 10.33 2.67 11 3.5 11H5" stroke="currentColor" strokeWidth="1.2"/>
                  </svg>
                )}
              </span>
            </div>

            <div className="gs-step">
              <span className="gs-step-num">2</span>
              <span className="gs-step-label">Tunnel</span>
            </div>
            <div className="gs-cmd-wrap gs-cmd-static">
              <code className="gs-cmd">phasethru http 3000</code>
            </div>

            <div className="gs-result">
              <span className="gs-arrow">→</span>
              <span>https://yourapp.phasethru.dev</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default GetStarted;
