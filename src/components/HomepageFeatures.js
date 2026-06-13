import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'Full MTProto 2.0',
    description: (
      <>
        Every one of the <strong>781 API methods</strong> is generated with embedded
        official docs. Call any of them directly through{' '}
        <Link to="/docs/basics/calling-the-api">
          <code>client.API()</code>
        </Link>
        .
      </>
    ),
  },
  {
    title: 'Users and bots',
    description: (
      <>
        Not just the Bot API. Sign in as a user with code, password,{' '}
        <Link to="/docs/authentication/qr-login">QR</Link> or{' '}
        <Link to="/docs/authentication/two-factor">2FA</Link>, or as a bot with a token.
      </>
    ),
  },
  {
    title: 'Batteries included',
    description: (
      <>
        High-level helpers for{' '}
        <Link to="/docs/helpers/message-sender">sending messages</Link>,{' '}
        <Link to="/docs/helpers/uploading-files">uploads</Link>,{' '}
        <Link to="/docs/helpers/downloading-files">downloads</Link> and{' '}
        <Link to="/docs/helpers/query-iterators">pagination</Link>.
      </>
    ),
  },
  {
    title: 'Fast and lightweight',
    description: (
      <ul>
        <li>~150&nbsp;KB per idle client</li>
        <li>No runtime reflection</li>
        <li>Thousands of concurrent clients</li>
      </ul>
    ),
  },
  {
    title: 'Robust',
    description: (
      <ul>
        <li>Automatic reconnects and DC migration</li>
        <li><Link to="/docs/helpers/updates-recovery">Update recovery</Link> engine</li>
        <li>Rigorously tested against real servers</li>
      </ul>
    ),
  },
  {
    title: 'Secure and capable',
    description: (
      <ul>
        <li>Follows Telegram security guidelines</li>
        <li><Link to="/docs/advanced/transports-and-proxy">MTProxy</Link>, WebSocket, WASM</li>
        <li>Voice and <Link to="/docs/advanced/calls">video calls</Link></li>
      </ul>
    ),
  },
];

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4', styles.feature)}>
      <div className="padding-horiz--md">
        <h3>{title}</h3>
        <div>{description}</div>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
