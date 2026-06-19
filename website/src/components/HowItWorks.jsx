import React from 'react';
import './HowItWorks.css';

const steps = [
  {
    num: '01',
    title: 'Run the binary',
    desc: 'Download a single Go binary. Point it at a local port.',
    code: '$ phasethru http 3000',
  },
  {
    num: '02',
    title: 'Tunnel connects',
    desc: 'An outbound TCP connection to the relay server is established — no inbound ports needed.',
    code: 'connected via relay',
  },
  {
    num: '03',
    title: 'Share your URL',
    desc: 'Your local service is now reachable at a public subdomain, instantly.',
    code: 'myapp.phasethru.dev',
  },
];

const HowItWorks = () => {
  return (
    <section id="how-it-works" className="hiw">
      <div className="wrap">
        <div className="hiw-label">How it works</div>
        <h2 className="hiw-title">Three steps. That's it.</h2>

        <div className="hiw-steps">
          {steps.map((s, i) => (
            <div key={s.num} className="hiw-step">
              <div className="hiw-step-header">
                <span className="hiw-num">{s.num}</span>
                {i < steps.length - 1 && (
                  <div className="hiw-connector">
                    <svg width="100%" height="2" fill="none">
                      <line x1="0" y1="1" x2="100%" y2="1" stroke="currentColor" strokeWidth="1" strokeDasharray="4 4" />
                    </svg>
                  </div>
                )}
              </div>
              <h3>{s.title}</h3>
              <p>{s.desc}</p>
              <div className="hiw-code">
                <code>{s.code}</code>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
