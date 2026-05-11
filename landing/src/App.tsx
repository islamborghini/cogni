import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeWindow } from "@/components/CodeWindow";
import { useEffect, useRef, useState } from "react";
import {
  ArrowRight,
  Check,
  Copy,
  Download,
  Github,
  Gauge,
  Lock,
  Layers,
  Workflow,
} from "lucide-react";

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const onClick = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // ignore
    }
  };
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={copied ? "Copied" : "Copy command"}
      className="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md border border-white/10 bg-black/60 text-zinc-400 hover:text-white hover:border-white/20"
    >
      {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  );
}

function Nav() {
  return (
    <header className="sticky top-0 z-30 w-full border-b border-white/10 bg-black/70 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <img src="/logo.svg" alt="cogni" className="h-7 w-7 rounded-md" />
          <span className="font-semibold tracking-tight">cogni</span>
        </div>
        <nav className="hidden items-center gap-6 text-sm text-zinc-400 md:flex">
          <a href="#problem" className="hover:text-white">
            Problem
          </a>
          <a href="#benefits" className="hover:text-white">
            Benefits
          </a>
          <a href="#benchmark" className="hover:text-white">
            Benchmark
          </a>
          <a href="#download" className="hover:text-white">
            Download
          </a>
          <a
            href="https://github.com/islamborghini/cogni"
            className="hover:text-white inline-flex items-center gap-1"
          >
            <Github className="h-4 w-4" /> GitHub
          </a>
        </nav>
        <a
          href="https://github.com/islamborghini/cogni"
          target="_blank"
          rel="noreferrer"
        >
          <Button size="sm">
            <Github className="h-4 w-4" />
            Star on GitHub
          </Button>
        </a>
      </div>
    </header>
  );
}

function Hero() {
  return (
    <section className="relative overflow-hidden">
      <div className="bg-grid absolute inset-0 opacity-60" />
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
      <div className="relative mx-auto grid max-w-6xl grid-cols-1 items-center gap-12 px-6 py-20 md:grid-cols-2 md:py-28">
        <div>
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-zinc-300">
            <span className="h-1.5 w-1.5 rounded-full bg-white" />
            Open core, MCP-native
          </div>
          <h1 className="text-6xl font-semibold tracking-tight md:text-8xl">
            Cogni
          </h1>
          <p className="mt-4 max-w-md text-lg text-zinc-400 md:text-xl">
            MCP to save your AI coding agent tokens
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <a href="#download">
              <Button size="lg">
                Get started
                <ArrowRight className="h-4 w-4" />
              </Button>
            </a>
            <a href="#problem">
              <Button variant="outline" size="lg">
                See how it works
              </Button>
            </a>
          </div>
          <div className="mt-10 inline-flex items-baseline gap-3 border-t border-white/10 pt-6">
            <span className="font-mono text-4xl font-semibold text-white">
              25%+
            </span>
            <span className="text-sm text-zinc-500">
              token reduction on real Python tasks
            </span>
          </div>
        </div>
        <div className="relative">
          <div className="absolute -inset-4 rounded-2xl bg-gradient-to-br from-white/10 to-transparent blur-2xl" />
          <CodeWindow className="relative" />
        </div>
      </div>
    </section>
  );
}

function DevinMark(_props: { className?: string }) {
  return (
    <span
      aria-label="Devin"
      className="font-semibold tracking-tight text-zinc-400 opacity-70 hover:opacity-100 transition-opacity text-base"
    >
      Devin
    </span>
  );
}

function OpenAIMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
      fill="currentColor"
      aria-label="OpenAI"
      className={className}
    >
      <path d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.91 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0 0 0 4.981 4.18a5.985 5.985 0 0 0-3.998 2.9 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0 .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0 5.772-4.206 5.99 5.99 0 0 0 3.997-2.9 6.056 6.056 0 0 0-.747-7.073zM13.26 22.43a4.476 4.476 0 0 1-2.876-1.04l.141-.081 4.779-2.758a.795.795 0 0 0 .392-.681v-6.737l2.02 1.168a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.494 4.494zM3.6 18.304a4.47 4.47 0 0 1-.535-3.014l.142.085 4.783 2.759a.771.771 0 0 0 .78 0l5.843-3.369v2.332a.08.08 0 0 1-.033.062L9.74 19.95a4.5 4.5 0 0 1-6.14-1.646zM2.34 7.896a4.485 4.485 0 0 1 2.366-1.973V11.6a.766.766 0 0 0 .388.676l5.815 3.355-2.02 1.168a.076.076 0 0 1-.071 0l-4.83-2.786A4.504 4.504 0 0 1 2.34 7.872zm16.597 3.855l-5.833-3.387L15.119 7.2a.076.076 0 0 1 .071 0l4.83 2.791a4.494 4.494 0 0 1-.676 8.105v-5.678a.79.79 0 0 0-.407-.667zm2.01-3.023l-.141-.085-4.774-2.782a.776.776 0 0 0-.785 0L9.409 9.23V6.897a.066.066 0 0 1 .028-.061l4.83-2.787a4.5 4.5 0 0 1 6.68 4.66zm-12.64 4.135l-2.02-1.164a.08.08 0 0 1-.038-.057V6.075a4.5 4.5 0 0 1 7.375-3.453l-.142.08-4.778 2.758a.795.795 0 0 0-.392.681zm1.097-2.365l2.602-1.5 2.607 1.5v2.999l-2.597 1.5-2.607-1.5z" />
    </svg>
  );
}

type Agent =
  | { name: string; slug: string }
  | { name: string; component: (props: { className?: string }) => JSX.Element };

const agents: Agent[] = [
  { name: "Claude", slug: "claude" },
  { name: "OpenAI", component: OpenAIMark },
  { name: "Gemini", slug: "googlegemini" },
  { name: "Cursor", slug: "cursor" },
  { name: "Windsurf", slug: "windsurf" },
  { name: "Devin", component: DevinMark },
  { name: "Zed", slug: "zedindustries" },
  { name: "GitHub Copilot", slug: "githubcopilot" },
  { name: "Replit", slug: "replit" },
];

const LogoRow = ({
  hidden,
  innerRef,
}: {
  hidden?: boolean;
  innerRef?: React.Ref<HTMLDivElement>;
}) => (
  <div
    ref={innerRef}
    aria-hidden={hidden}
    className="flex flex-shrink-0 items-center gap-10 pr-10"
  >
    {agents.map((a) => {
      const iconCls =
        "h-7 w-7 text-zinc-400 opacity-70 hover:opacity-100 transition-opacity";
      const box = "flex h-7 w-20 items-center justify-center flex-shrink-0";
      if ("component" in a) {
        const Cmp = a.component;
        return (
          <div key={a.name} className={box}>
            <Cmp className={iconCls} />
          </div>
        );
      }
      return (
        <div key={a.name} className={box}>
          <img
            src={`https://cdn.simpleicons.org/${a.slug}/a1a1aa`}
            alt={a.name}
            title={a.name}
            loading="eager"
            decoding="sync"
            className="h-7 w-7 opacity-70 hover:opacity-100 transition-opacity"
          />
        </div>
      );
    })}
  </div>
);

function AgentsStrip() {
  const rowRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(0);

  useEffect(() => {
    const el = rowRef.current;
    if (!el) return;
    const measure = () => setWidth(el.offsetWidth);
    measure();
    const ro = new ResizeObserver(measure);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  // 35px/sec for a relaxed scroll pace.
  const duration = width > 0 ? `${width / 35}s` : "45s";

  return (
    <section className="border-y border-white/10 bg-black">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <p className="text-center font-mono text-xs uppercase tracking-widest text-zinc-500">
          Use it with your favourite agents
        </p>
        <div className="marquee-mask relative mt-8 overflow-hidden">
          <div
            className="flex animate-marquee"
            style={
              {
                "--marquee-width": width ? `${width}px` : "50%",
                "--marquee-duration": duration,
              } as React.CSSProperties
            }
          >
            <LogoRow innerRef={rowRef} />
            <LogoRow hidden />
          </div>
        </div>
      </div>
    </section>
  );
}

function Problem() {
  return (
    <section id="problem" className="border-y border-white/10 bg-black">
      <div className="mx-auto max-w-6xl px-6 py-20">
        <div className="grid gap-10 md:grid-cols-2">
          <div>
            <p className="font-mono text-xs uppercase tracking-widest text-zinc-500">
              The problem
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-tight md:text-4xl">
              Coding agents forget. Then they pay to remember.
            </h2>
            <p className="mt-5 text-zinc-400 leading-relaxed">
              Every prompt forces agents to rediscover structure they have
              already seen. Read, Grep, Glob, repeat. Developers hit usage
              limits. Teams burn budget on tokens that produced no new work.
            </p>
            <p className="mt-4 text-zinc-400 leading-relaxed">
              Cogni indexes the repo once and exposes structured recall over
              MCP. The agent asks for a symbol and gets the symbol, not the
              file.
            </p>
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Card>
              <CardHeader>
                <Gauge className="h-5 w-5 text-white" />
                <CardTitle className="mt-3">Measured savings</CardTitle>
              </CardHeader>
              <CardContent>
                25%+ mean token reduction on Python tasks across the httpx
                codebase, with no drop in pass rate.
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Layers className="h-5 w-5 text-white" />
                <CardTitle className="mt-3">Structured recall</CardTitle>
              </CardHeader>
              <CardContent>
                Five MCP tools: repo overview, file outline, symbol search,
                symbol source, find references.
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Workflow className="h-5 w-5 text-white" />
                <CardTitle className="mt-3">Drop-in for any agent</CardTitle>
              </CardHeader>
              <CardContent>
                Works with Claude Code, Cursor, or any MCP-aware client. One
                config line, one binary.
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Lock className="h-5 w-5 text-white" />
                <CardTitle className="mt-3">Local first</CardTitle>
              </CardHeader>
              <CardContent>
                Index lives at ~/.cogni. No code leaves your machine. Open
                source under Apache 2.0.
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </section>
  );
}

const benefitItems = [
  {
    k: "01",
    h: "Lower spend per task",
    p: "Fewer input tokens per agent loop. The savings compound across long sessions and large codebases.",
  },
  {
    k: "02",
    h: "Faster iteration",
    p: "Structured queries return in milliseconds. The agent stops paginating through files it does not need.",
  },
  {
    k: "03",
    h: "Reproducible benchmarks",
    p: "Pinned commits, fixed task set, public methodology. Run it yourself, on your own repos.",
  },
  {
    k: "04",
    h: "Open core",
    p: "Apache 2.0 today. Self-host the binary, integrate with your stack, contribute upstream.",
  },
];

function Benefits() {
  return (
    <section id="benefits" className="border-t border-white/10 bg-black">
      <div className="mx-auto w-full max-w-6xl px-6 py-20">
        <p className="font-mono text-xs uppercase tracking-widest text-zinc-500">
          Why cogni
        </p>
        <h2 className="mt-3 text-3xl font-semibold tracking-tight md:text-4xl">
          Get more out of your AI coding agent.
        </h2>

        <div className="mt-10 grid grid-cols-1 gap-px overflow-hidden rounded-xl border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-4">
          {benefitItems.map((it) => (
            <div key={it.k} className="bg-black p-6">
              <div className="font-mono text-xs text-zinc-600">{it.k}</div>
              <h3 className="mt-3 text-base font-semibold text-zinc-200">
                {it.h}
              </h3>
              <p className="mt-2 text-sm text-zinc-400 leading-relaxed">
                {it.p}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function Benchmark() {
  return (
    <section
      id="benchmark"
      className="border-t border-white/10 bg-black"
    >
      <div className="mx-auto w-full max-w-6xl px-6 py-20">
        <p className="font-mono text-xs uppercase tracking-widest text-zinc-500">
          Benchmark
        </p>
        <h2 className="mt-3 text-3xl font-semibold tracking-tight md:text-4xl">
          Measured on a real Python codebase.
        </h2>
        <p className="mt-3 max-w-2xl text-zinc-400 leading-relaxed">
          We run a fixed task against a pinned commit of{" "}
          <code className="font-mono text-zinc-300">httpx</code> with and
          without Cogni registered. Five runs per condition. Tokens come from
          the Claude Code SDK, not estimates.
        </p>

        <div className="mt-10 grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle className="font-mono text-xs uppercase tracking-widest text-zinc-500">
                Token reduction
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="font-mono text-4xl font-semibold text-white">
                24.3%
              </div>
              <p className="mt-2 text-sm text-zinc-500">
                2,209 baseline tokens dropped to 1,673 with Cogni.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="font-mono text-xs uppercase tracking-widest text-zinc-500">
                Pass rate
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="font-mono text-4xl font-semibold text-white">
                100%
              </div>
              <p className="mt-2 text-sm text-zinc-500">
                Same as the baseline. No regression in answer quality.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="font-mono text-xs uppercase tracking-widest text-zinc-500">
                Task
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="font-mono text-lg font-semibold text-white">
                explain-transports
              </div>
              <p className="mt-2 text-sm text-zinc-500">
                Family: explain. Repo: httpx. Runs: 5 per condition.
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="mt-8 overflow-x-auto rounded-xl border border-white/10">
          <table className="w-full text-left text-sm">
            <thead className="bg-white/5 text-zinc-400">
              <tr>
                <th className="px-4 py-3 font-medium">Condition</th>
                <th className="px-4 py-3 font-medium text-right">Mean tokens</th>
                <th className="px-4 py-3 font-medium text-right">Pass rate</th>
              </tr>
            </thead>
            <tbody className="font-mono text-zinc-200">
              <tr className="border-t border-white/10">
                <td className="px-4 py-3">Baseline (Read / Grep / Glob)</td>
                <td className="px-4 py-3 text-right">2,209</td>
                <td className="px-4 py-3 text-right">100%</td>
              </tr>
              <tr className="border-t border-white/10">
                <td className="px-4 py-3">With Cogni</td>
                <td className="px-4 py-3 text-right text-white">1,673</td>
                <td className="px-4 py-3 text-right">100%</td>
              </tr>
            </tbody>
          </table>
        </div>

        <p className="mt-6 text-sm text-zinc-500">
          <a
            href="https://github.com/islamborghini/cogni/blob/main/BENCHMARK.md"
            className="underline hover:text-zinc-300"
          >
            Read more
          </a>
        </p>
      </div>
    </section>
  );
}

const installSteps: Array<{ k: string; h: string; cmd?: string; note?: string }> = [
  {
    k: "01",
    h: "Install the binary",
    cmd: "brew install islamborghini/tap/cogni",
    note: "macOS and Linux. Static binary, no runtime.",
  },
  {
    k: "02",
    h: "Run install in your repo",
    cmd: "cd path/to/your/python/repo && cogni install",
    note: "Registers the MCP server, writes CLAUDE.md, indexes the repo. Idempotent.",
  },
  {
    k: "03",
    h: "Restart Claude Code",
    note: "Open Claude Code in the repo. Ask: \"use repo_overview to explain this codebase\" to verify.",
  },
];

function DownloadSection() {
  return (
    <section
      id="download"
      className="relative overflow-hidden border-t border-white/10"
    >
      <div className="bg-grid absolute inset-0 opacity-40" />
      <div className="relative mx-auto max-w-4xl px-6 py-24">
        <div className="text-center">
          <p className="font-mono text-xs uppercase tracking-widest text-zinc-500">
            Quickstart
          </p>
          <h2 className="mt-3 text-3xl font-semibold tracking-tight md:text-5xl">
            Three commands. Then your agent is faster.
          </h2>
          <p className="mx-auto mt-4 max-w-xl text-zinc-400">
            Cogni is a single static binary that registers itself with Claude Code in one step. Nothing leaves your machine.
          </p>
        </div>

        <ol className="mt-10 space-y-3">
          {installSteps.map((s) => (
            <li
              key={s.k}
              className="rounded-xl border border-white/10 bg-[#050505] p-5"
            >
              <div className="flex items-baseline gap-3">
                <span className="font-mono text-xs text-green-700 tracking-wider">
                  {s.k}
                </span>
                <h3 className="text-base font-semibold text-white">{s.h}</h3>
              </div>
              {s.cmd && (
                <div className="relative mt-3">
                  <pre className="overflow-x-auto rounded-md border border-white/5 bg-black px-4 py-3 pr-12 font-mono text-sm text-zinc-200">
                    <code>
                      <span className="text-zinc-600">$ </span>
                      {s.cmd}
                    </code>
                  </pre>
                  <CopyButton text={s.cmd} />
                </div>
              )}
              {s.note && (
                <p className="mt-3 text-sm text-zinc-500">{s.note}</p>
              )}
            </li>
          ))}
        </ol>

        <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
          <a
            href="https://github.com/islamborghini/cogni"
            target="_blank"
            rel="noreferrer"
          >
            <Button size="lg">
              <Github className="h-4 w-4" />
              Star on GitHub
            </Button>
          </a>
          <a href="https://github.com/islamborghini/cogni/releases/latest">
            <Button variant="outline" size="lg">
              <Download className="h-4 w-4" />
              Latest release
            </Button>
          </a>
          <a href="https://github.com/islamborghini/cogni#readme">
            <Button variant="ghost" size="lg">
              Docs
              <ArrowRight className="h-4 w-4" />
            </Button>
          </a>
        </div>

        <p className="mt-6 text-center text-xs text-zinc-600">
          Need manual setup? See the{" "}
          <a
            href="https://github.com/islamborghini/cogni#manual-configuration"
            className="underline hover:text-zinc-400"
          >
            manual configuration guide
          </a>
          .
        </p>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer className="border-t border-white/10 bg-black">
      <div className="mx-auto flex max-w-6xl flex-col items-start justify-between gap-4 px-6 py-10 sm:flex-row sm:items-center">
        <div className="flex items-center gap-2">
          <img src="/logo.svg" alt="cogni" className="h-6 w-6 rounded-md" />
          <span className="font-semibold">cogni</span>
          <span className="ml-3 text-xs text-zinc-500">
            Structured recall for coding agents.
          </span>
        </div>
        <div className="flex items-center gap-6 text-sm text-zinc-500">
          <a href="#problem" className="hover:text-white">
            Problem
          </a>
          <a href="#benefits" className="hover:text-white">
            Benefits
          </a>
          <a href="#benchmark" className="hover:text-white">
            Benchmark
          </a>
          <a href="#download" className="hover:text-white">
            Download
          </a>
          <a
            href="https://github.com/islamborghini/cogni"
            className="hover:text-white"
          >
            GitHub
          </a>
        </div>
        <div className="text-xs text-zinc-600">
          © {new Date().getFullYear()} cogni · Apache 2.0
        </div>
      </div>
    </footer>
  );
}

export default function App() {
  return (
    <div className="min-h-full font-sans">
      <Nav />
      <main>
        <Hero />
        <AgentsStrip />
        <Problem />
        <Benefits />
        <Benchmark />
        <DownloadSection />
      </main>
      <Footer />
    </div>
  );
}
