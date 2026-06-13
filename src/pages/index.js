import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import CodeBlock from '@theme/CodeBlock';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';

const quickStart = `package main

import (
	"context"

	"github.com/gotd/td/telegram"
)

func main() {
	// Get api_id and api_hash from https://my.telegram.org/apps.
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// The client is connected only while this function runs.
		api := client.API()

		// Call any of the 781 MTProto methods directly.
		dc, err := api.HelpGetNearestDC(ctx)
		if err != nil {
			return err
		}
		_ = dc
		return nil
	}); err != nil {
		panic(err)
	}
}`;

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <p className={styles.heroDescription}>
          A full MTProto 2.0 API client in pure Go for <strong>users and bots</strong> —
          the same protocol the official apps and TDLib speak, with batteries-included
          helpers for auth, uploads, downloads and updates.
        </p>
        <div className={styles.install}>
          <CodeBlock language="bash">go get github.com/gotd/td@latest</CodeBlock>
        </div>
        <div className={styles.buttons}>
          <Link className="button button--secondary button--lg" to="/docs/intro">
            Get started
          </Link>
          <Link className="button button--secondary button--lg" to="/docs/reference">
            API Reference
          </Link>
          <Link
            className="button button--outline button--secondary button--lg"
            href="https://github.com/gotd/td">
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

function QuickStart() {
  return (
    <section className={styles.quickStart}>
      <div className="container">
        <div className="row">
          <div className="col col--8 col--offset-2">
            <h2 className={styles.sectionTitle}>Connect in a few lines</h2>
            <CodeBlock language="go">{quickStart}</CodeBlock>
            <p className={styles.quickStartFootnote}>
              See the <Link to="/docs/getting-started/first-client">first-client guide</Link> and{' '}
              <Link to="/docs/basics/echo-bot">echo bot tutorial</Link> to go further.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Telegram MTProto API client in Go for users and bots.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <QuickStart />
      </main>
    </Layout>
  );
}
