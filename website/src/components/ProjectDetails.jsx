import React from 'react';
import './ProjectDetails.css';

const ProjectDetails = () => {
  return (
    <section id="features" className="features">
      <div className="wrap">
        <div className="feat-top">
          <div className="feat-top-left">
            <div className="feat-label">Features</div>
            <h2>Designed around<br />how developers actually work.</h2>
          </div>
          <p className="feat-top-right">
            No agents running in the background, no YAML to wrangle,
            no accounts to create for local testing. PhaseThru gets out
            of your way.
          </p>
        </div>

        <div className="feat-grid">
          <div className="feat-card">
            <div className="feat-icon-wrap">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
              </svg>
            </div>
            <h3>Yamux Multiplexing</h3>
            <p>
              Hundreds of concurrent HTTP streams flow over a single
              persistent TCP connection. No reconnection churn, no
              wasted sockets, sub-millisecond overhead.
            </p>
          </div>

          {/* Small features */}
          <div className="feat-card">
            <div className="feat-icon-wrap">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </div>
            <h3>NAT Piercing</h3>
            <p>Outbound-only TCP bypasses even CGNAT and enterprise firewalls.</p>
          </div>

          <div className="feat-card">
            <div className="feat-icon-wrap">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
              </svg>
            </div>
            <h3>Single Binary</h3>
            <p>Statically compiled Go. Drop it on any box and run — zero dependencies.</p>
          </div>

          <div className="feat-card">
            <div className="feat-icon-wrap">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </div>
            <h3>In-Memory TLS</h3>
            <p>Certificates generated at startup, held in RAM. No files on disk.</p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ProjectDetails;
