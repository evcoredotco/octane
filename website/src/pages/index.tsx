import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import styles from './index.module.css';

type ResultRow = {
  sym: string;
  symClass: string;
  name: string;
  status: string;
  statusClass: string;
  time: string;
};

const RESULTS: ResultRow[] = [
  {sym: '✓', symClass: styles.ok, name: 'station_connection_established', status: 'PASS', statusClass: styles.pass, time: '11ms'},
  {sym: '✓', symClass: styles.ok, name: 'station_boot_accepted', status: 'PASS', statusClass: styles.pass, time: '44ms'},
  {sym: '✓', symClass: styles.ok, name: 'boot_sequence_accepted', status: 'PASS', statusClass: styles.pass, time: '82ms'},
  {sym: '✓', symClass: styles.ok, name: 'connector_status_available', status: 'PASS', statusClass: styles.pass, time: '19ms'},
  {sym: '✓', symClass: styles.ok, name: 'transaction_pluginfirst_accepted', status: 'PASS', statusClass: styles.pass, time: '91ms'},
  {sym: '↺', symClass: styles.cached, name: 'meter_values_periodic_accepted', status: 'CACHED', statusClass: styles.cached, time: '0ms'},
  {sym: '✓', symClass: styles.ok, name: 'connector_reservation_faulted', status: 'PASS', statusClass: styles.pass, time: '37ms'},
];

function pad(value: string, width: number): string {
  return value.length >= width ? value + ' ' : value.padEnd(width);
}

function Terminal(): ReactNode {
  return (
    <div className={styles.terminal}>
      <div className={styles.termBar}>
        <span className={`${styles.dot} ${styles.dotR}`} />
        <span className={`${styles.dot} ${styles.dotY}`} />
        <span className={`${styles.dot} ${styles.dotG}`} />
        <span className={styles.termTitle}>octane@csms — conformance run</span>
      </div>
      <code className={styles.termBody}>
        <div>
          <span className={styles.prompt}>$ </span>
          <span className={styles.cmd}>octane run scenarios/v16 --csms-endpoint ws://localhost:9210</span>
        </div>
        <div>
          <span className={styles.muted}>resolving dependency graph · 21 stories · topological order</span>
        </div>
        {RESULTS.map((r) => (
          <div key={r.name}>
            <span className={r.symClass}>{r.sym}</span>
            {' ' + pad(r.name, 34)}
            <span className={r.statusClass}>{pad(r.status, 8)}</span>
            <span className={styles.muted}>{r.time}</span>
          </div>
        ))}
        <div>
          <span className={styles.muted}>  … 14 more stories</span>
        </div>
        <div>{' '}</div>
        <div>
          <span className={styles.summary}>passed=</span>
          <span className={styles.pass}>21</span>
          <span className={styles.summary}> failed=</span>
          <span className={styles.pass}>0</span>
          <span className={styles.summary}> skipped=</span>
          <span className={styles.pass}>0</span>
          <span className={styles.summary}> cache-hits=</span>
          <span className={styles.cached}>6</span>
        </div>
        <div>
          <span className={styles.muted}>report-dir=</span>
          <span className={styles.path}>reports/run-20260628-1/</span>
          <span className={styles.cursor} />
        </div>
      </code>
    </div>
  );
}

function Hero(): ReactNode {
  return (
    <header className={styles.hero}>
      <div className={styles.heroGrid} />
      <div className={styles.heroInner}>
        <div>
          <span className={styles.eyebrow}>▸ OCPP 1.6J conformance</span>
          <h1 className={styles.title}>
            Prove your CSMS speaks OCPP <span className={styles.titleAccent}>to the letter.</span>
          </h1>
          <p className={styles.subtitle}>
            OCTANE impersonates charging stations over the OCPP-J WebSocket and
            asserts your Charging Station Management System answers exactly as the
            specification requires — at the wire, with no changes to the system under test.
          </p>
          <div className={styles.ctaRow}>
            <Link className={styles.btnPrimary} to="/docs/getting-started">
              Get started →
            </Link>
            <Link className={styles.btnGhost} to="/docs/intro">
              How it works
            </Link>
            <Link className={styles.btnGhost} href="https://github.com/evcoreco/octane">
              ★ GitHub
            </Link>
          </div>
          <p className={styles.installLine}>
            <span className={styles.prompt}>$</span>
            <code>go build ./cmd/octane</code>
            build from source · Go 1.26+
          </p>
        </div>
        <Terminal />
      </div>
    </header>
  );
}

type Feature = {index: string; title: string; body: string};

const FEATURES: Feature[] = [
  {
    index: '01',
    title: 'Wire-level conformance',
    body: 'OCTANE tests only what crosses the WebSocket. No admin APIs, no sidecar service, no changes to the CSMS under test. A deviation is a finding, never a tolerance.',
  },
  {
    index: '02',
    title: 'Declarative .story DSL',
    body: 'Scenarios are Gherkin-flavored text files that read like plain English and trace directly to the OCPP spec section each one exercises.',
  },
  {
    index: '03',
    title: 'Dependency graph + cache',
    body: 'Stories declare prerequisites. The runner resolves a DAG, runs them in order, and a content-addressed cache skips anything unchanged since the last run.',
  },
  {
    index: '04',
    title: 'Two surfaces, one engine',
    body: 'Everything reachable from the octane CLI is reachable from the octane-action GitHub Action, and vice versa. Same binary, same behavior.',
  },
  {
    index: '05',
    title: 'Deterministic reports',
    body: 'Each run emits a byte-stable report.json plus a Robot Framework output.xml from a single tree — ready for Allure, ReportPortal, Jenkins, GitLab, and Actions.',
  },
  {
    index: '06',
    title: '17 OCPP 1.6 message types',
    body: 'Boot and heartbeat, transactions, remote control, configuration, availability, and reservations are covered today — with more on the way.',
  },
];

function Features(): ReactNode {
  return (
    <section className={styles.section}>
      <div className={styles.sectionInner}>
        <p className={styles.kicker}>// why octane</p>
        <h2 className={styles.sectionTitle}>Conformance you can trust, in CI you already run</h2>
        <p className={styles.sectionLede}>
          OCTANE was built on one principle: the conformance signal must be honest.
          It observes the protocol from the charging-station side and refuses every
          shortcut that would let a non-conformant CSMS pass.
        </p>
        <div className={styles.features}>
          {FEATURES.map((f) => (
            <div className={styles.card} key={f.index}>
              <div className={styles.cardIndex}>{f.index}</div>
              <h3 className={styles.cardTitle}>{f.title}</h3>
              <p className={styles.cardBody}>{f.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

type LayerDef = {tag: string; name: string; desc: string};

const LAYERS: LayerDef[] = [
  {
    tag: 'Layer 1',
    name: 'Stories',
    desc: '.story files — one scenario each, version-controlled, and traceable to an OCPP specification section via Spec-Ref.',
  },
  {
    tag: 'Layer 2',
    name: 'Keywords',
    desc: 'Typed Go functions that map step text to wire actions: domain keywords (OCPP 1.6 semantics) resolved over a primitive transport layer.',
  },
  {
    tag: 'Layer 3',
    name: 'Engine',
    desc: 'WebSocket transport, OCPP-J framing, the DAG runner and worker pool, the content-addressed cache, and a deterministic clock and RNG.',
  },
];

function HowItWorks(): ReactNode {
  return (
    <section className={`${styles.section} ${styles.sectionAlt}`}>
      <div className={styles.sectionInner}>
        <p className={styles.kicker}>// architecture</p>
        <h2 className={styles.sectionTitle}>Three layers, one contract each</h2>
        <p className={styles.sectionLede}>
          Stories never import Go. Keywords never know which CSMS they are talking to.
          The engine never knows which OCPP version it is driving. Each layer is the
          contract for the one above it.
        </p>
        <div className={styles.layers}>
          <div className={styles.layer}>
            <span className={styles.layerTag}>{LAYERS[0].tag}</span>
            <div>
              <p className={styles.layerName}>{LAYERS[0].name}</p>
              <p className={styles.layerDesc}>{LAYERS[0].desc}</p>
            </div>
          </div>
          <p className={styles.layerArrow}>▼ resolves keywords against</p>
          <div className={styles.layer}>
            <span className={styles.layerTag}>{LAYERS[1].tag}</span>
            <div>
              <p className={styles.layerName}>{LAYERS[1].name}</p>
              <p className={styles.layerDesc}>{LAYERS[1].desc}</p>
            </div>
          </div>
          <p className={styles.layerArrow}>▼ drives the wire via</p>
          <div className={styles.layer}>
            <span className={styles.layerTag}>{LAYERS[2].tag}</span>
            <div>
              <p className={styles.layerName}>{LAYERS[2].name}</p>
              <p className={styles.layerDesc}>{LAYERS[2].desc}</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

const ASSERTIONS: string[] = [
  'BootNotification is answered with status "Accepted" within 30 seconds',
  'The advertised heartbeatInterval falls between 30 and 86400 seconds',
  'currentTime is returned as valid ISO-8601',
  'The CSMS acknowledges each Heartbeat it advertised the cadence for',
];

function Showcase(): ReactNode {
  return (
    <section className={styles.section}>
      <div className={styles.sectionInner}>
        <p className={styles.kicker}>// the story is the test</p>
        <h2 className={styles.sectionTitle}>Readable scenarios. Wire-level assertions.</h2>
        <p className={styles.sectionLede}>
          A reviewer reads the same file the runner executes. Every step maps to a
          keyword that sends an OCPP message and checks the response.
        </p>
        <div className={styles.showcase}>
          <div className={styles.codePanel}>
            <code className={styles.codePanelBody}>
              <span className={styles.key}>Meta</span>{'\n'}
              {'    '}<span className={styles.key}>Id:</span>{'        boot_sequence_accepted\n'}
              {'    '}<span className={styles.key}>Spec-Ref:</span>{'  OCPP-J 1.6 §6.5 BootNotification\n'}
              {'    '}<span className={styles.key}>Stations:</span>{'  1\n'}
              {'    '}<span className={styles.key}>Depends:</span>{'\n'}
              {'      - id: station_connection_established\n'}
              {'\n'}
              <span className={styles.kw}>Scenario</span>{': Cold-boot registration sequence\n'}
              {'    '}<span className={styles.kw}>When</span>{'  station '}<span className={styles.str}>"CP01"</span>{' sends BootNotification\n'}
              {'          with reason '}<span className={styles.str}>"PowerUp"</span>{'\n'}
              {'    '}<span className={styles.kw}>Then</span>{'  the CSMS responds with status '}<span className={styles.str}>"Accepted"</span>{'\n'}
              {'          within 30 seconds\n'}
              {'    '}<span className={styles.kw}>And</span>{'   the response includes a heartbeatInterval\n'}
              {'          between 30 and 86400\n'}
            </code>
          </div>
          <ul className={styles.assertList}>
            {ASSERTIONS.map((a) => (
              <li className={styles.assertItem} key={a}>
                <span className={styles.assertCheck}>✓</span>
                <span>{a}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}

type Stat = {num: string; label: string};

const STATS: Stat[] = [
  {num: '17', label: 'OCPP 1.6 message types covered'},
  {num: '21', label: 'conformance & helper stories'},
  {num: '2', label: 'report formats — JSON + Robot XML'},
  {num: '0', label: 'changes required to your CSMS'},
];

function Stats(): ReactNode {
  return (
    <section className={`${styles.section} ${styles.sectionAlt}`}>
      <div className={styles.sectionInner}>
        <div className={styles.stats}>
          {STATS.map((s) => (
            <div key={s.label}>
              <div className={styles.statNum}>{s.num}</div>
              <div className={styles.statLabel}>{s.label}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function FinalCta(): ReactNode {
  return (
    <section className={styles.ctaBand}>
      <h2 className={styles.ctaTitle}>Gate every commit on OCPP conformance.</h2>
      <p className={styles.ctaText}>
        Point OCTANE at a CSMS endpoint, run the suite locally or in CI, and read a
        deterministic report. No mocks, no vendor adapters, no excuses.
      </p>
      <div className={styles.ctaRow} style={{justifyContent: 'center'}}>
        <Link className={styles.btnPrimary} to="/docs/getting-started">
          Get started →
        </Link>
        <Link className={styles.btnGhost} to="/docs/concepts/wire-conformance">
          Read the concepts
        </Link>
      </div>
      <div>
        <span className={styles.noteBadge}>pre-alpha · build from source · Apache-2.0</span>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  return (
    <Layout
      title="OCPP 1.6J conformance testing"
      description="OCTANE is an open-source conformance harness for OCPP 1.6J. It impersonates charging stations over the wire and verifies your CSMS responds to spec — no CSMS changes required."
    >
      <div className={styles.page}>
        <Hero />
        <main>
          <Features />
          <HowItWorks />
          <Showcase />
          <Stats />
          <FinalCta />
        </main>
      </div>
    </Layout>
  );
}
