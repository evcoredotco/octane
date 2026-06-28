import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';

export default function Home(): JSX.Element {
  return (
    <Layout
      title="OCTANE"
      description="OCPP Conformance Testing and Network Evaluation"
    >
      <main className="container margin-vert--xl">
        <h1>OCTANE</h1>
        <p>
          OCPP Conformance Testing and Network Evaluation for CSMS teams that
          need repeatable wire-level checks.
        </p>
        <p>
          <Link className="button button--primary" to="/docs/intro">
            Read the docs
          </Link>
        </p>
      </main>
    </Layout>
  );
}

