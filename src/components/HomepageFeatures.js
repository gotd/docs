import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'Fast',
    description: (
        <ul>
            <li>Code generated MTProto</li>
            <li>Low memory overhead, 150kb per idle client</li>
            <li>No runtime reflection</li>
        </ul>
    ),
  },
  {
    title: 'Robust',
    description: (
        <ul>
            <li>Automatic re-connects with keepalive</li>
            <li>Rigorously tested</li>
        </ul>
    ),
  },
  {
    title: 'Feature-rich',
    description: (
        <ul>
            <li>2FA</li>
            <li>MTProxy</li>
            <li>Websocket and WASM support</li>
            <li><code>context.Context</code> aware</li>
        </ul>
    ),
  },
];

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
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
