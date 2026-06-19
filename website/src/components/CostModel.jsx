import React from 'react';
import { Link } from 'react-router-dom';
import './CostModel.css';

const tiers = [
  {
    name: 'Free',
    price: '0',
    tagline: 'For tinkering and side projects.',
    features: [
      '1 tunnel',
      'Random subdomain',
      'Community support',
      'HTTP & HTTPS proxying',
    ],
    cta: 'Start for Free',
    ctaClass: 'btn-ghost',
    href: '/auth',
  },
  {
    name: 'Pro',
    price: '15',
    tagline: 'For teams shipping to production.',
    badge: 'Popular',
    features: [
      'Unlimited tunnels',
      'Custom domains (BYOD)',
      'Priority support',
      'Global edge network',
      'Team seats & management',
    ],
    cta: 'Start Free Trial',
    ctaClass: 'btn-fill',
    href: '/auth',
  },
];

const CostModel = () => {
  return (
    <section id="pricing" className="pricing">
      <div className="wrap">
        <div className="pricing-top">
          <div className="pricing-label">Pricing</div>
          <h2>Simple pricing,<br />no surprises.</h2>
          <p>Start free, scale when you need to. Cancel anytime.</p>
        </div>
        <div className="pricing-cards">
          {tiers.map((t) => (
            <div key={t.name} className={`p-card ${t.badge ? 'p-card--feat' : ''}`}>
              {t.badge && <span className="p-badge">{t.badge}</span>}
              <div className="p-card-top">
                <h3>{t.name}</h3>
                <div className="p-price">
                  <span className="p-dollar">$</span>
                  <span className="p-amount">{t.price}</span>
                  <span className="p-period">/mo</span>
                </div>
                <p className="p-tagline">{t.tagline}</p>
              </div>
              <ul className="p-features">
                {t.features.map((f) => (
                  <li key={f}>
                    <svg width="15" height="15" viewBox="0 0 16 16" fill="none">
                      <path d="M3.5 8.5L6.5 11.5L12.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    {f}
                  </li>
                ))}
              </ul>
              <Link to={t.href} className={t.ctaClass}>{t.cta}</Link>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default CostModel;
