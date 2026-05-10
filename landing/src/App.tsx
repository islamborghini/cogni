import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeWindow } from "@/components/CodeWindow";
import {
  ArrowRight,
  Download,
  Github,
  Gauge,
  Lock,
  Layers,
  Workflow,
} from "lucide-react";

function Nav() {
  return (
    <header className="sticky top-0 z-30 w-full border-b border-white/10 bg-black/70 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <div className="h-6 w-6 rounded-md border border-white/15 bg-white/5 grid place-items-center">
            <span className="font-mono text-[11px]">c</span>
          </div>
          <span className="font-semibold tracking-tight">cogni</span>
          <span className="ml-3 hidden rounded-full border border-white/10 px-2 py-0.5 text-[10px] font-mono uppercase tracking-wider text-zinc-400 sm:inline">
            v0.1.1
          </span>
        </div>
        <nav className="hidden items-center gap-6 text-sm text-zinc-400 md:flex">
          <a href="#problem" className="hover:text-white">
            Problem
          </a>
          <a href="#benefits" className="hover:text-white">
            Benefits
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
        <a href="#download">
          <Button size="sm">Get cogni</Button>
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
          <h1 className="text-5xl font-semibold tracking-tight md:text-6xl">
            Cogni
          </h1>
          <p className="mt-4 max-w-md text-lg text-zinc-300 md:text-xl">
            Structured recall that cuts coding-agent token use.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <a href="#download">
              <Button size="lg">
                <Download className="h-4 w-4" />
                Download
              </Button>
            </a>
            <a href="#problem">
              <Button variant="outline" size="lg">
                See how it works
                <ArrowRight className="h-4 w-4" />
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
              Coding agents brute-force context.
            </h2>
            <p className="mt-5 text-zinc-400 leading-relaxed">
              Every Read, Grep, and Glob round-trips raw source through the
              model. On a real repository, that means thousands of tokens
              spent re-reading files the agent has already seen, just to find
              one function. You pay for it. The agent gets slower. The answer
              gets worse.
            </p>
            <p className="mt-4 text-zinc-400 leading-relaxed">
              Cogni indexes the repo once, then exposes structured recall
              over MCP. The agent asks for a symbol; it gets the symbol, not
              the file.
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
  const sectionRef = useRef<HTMLElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const cardRefs = useRef<Array<HTMLDivElement | null>>([]);
  const [active, setActive] = useState(0);
  const [pos, setPos] = useState<{ x: number; y: number; visible: boolean }>({
    x: 0,
    y: 0,
    visible: false,
  });

  useEffect(() => {
    let ticking = false;
    const N = benefitItems.length;

    const compute = () => {
      ticking = false;
      const sec = sectionRef.current;
      const cont = containerRef.current;
      if (!sec || !cont) return;
      const secRect = sec.getBoundingClientRect();
      const contRect = cont.getBoundingClientRect();
      const vh = window.innerHeight;

      // progress: 0 when section top hits viewport center, 1 when section bottom hits it.
      const total = secRect.height;
      const passed = vh / 2 - secRect.top;
      let p = total > 0 ? passed / total : 0;
      // squeeze active range into the middle 70% so the user actually sees
      // the first and last cards highlighted at the section's edges.
      const lo = 0.35;
      const hi = 0.65;
      const t = Math.min(1, Math.max(0, (p - lo) / (hi - lo)));

      // continuous index in [0, N-1] for smooth interpolation
      const continuous = t * (N - 1);
      const idx = Math.round(continuous);
      const lower = Math.floor(continuous);
      const upper = Math.min(N - 1, lower + 1);
      const localT = continuous - lower;

      // interpolate x between adjacent card centers, y between their tops
      const a = cardRefs.current[lower];
      const b = cardRefs.current[upper];
      if (a && b) {
        const ar = a.getBoundingClientRect();
        const br = b.getBoundingClientRect();
        const ax = ar.left - contRect.left + ar.width / 2;
        const bx = br.left - contRect.left + br.width / 2;
        const ay = ar.top - contRect.top;
        const by = br.top - contRect.top;
        const x = ax + (bx - ax) * localT;
        const y = ay + (by - ay) * localT;
        const onScreen = secRect.top < vh && secRect.bottom > 0;
        setPos({ x, y, visible: onScreen });
      }
      setActive(idx);
    };

    const onScroll = () => {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(compute);
    };

    compute();
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("resize", onScroll);
    return () => {
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("resize", onScroll);
    };
  }, []);

  return (
    <section
      id="benefits"
      ref={sectionRef}
      className="relative bg-black"
      style={{ height: "150vh" }}
    >
      <div className="sticky top-20">
        <div className="mx-auto w-full max-w-6xl px-6 py-6">
          <p className="font-mono text-xs uppercase tracking-widest text-zinc-500">
            Why cogni
          </p>
          <h2 className="mt-3 text-3xl font-semibold tracking-tight md:text-4xl">
            Built for the people paying the bill.
          </h2>
          <p className="mt-3 text-sm text-zinc-500">
            Scroll to walk through each benefit.
          </p>

          <div ref={containerRef} className="relative mt-8">
          {/* floating thick green dollar that smoothly tracks the active card */}
          <div
            aria-hidden
            className="pointer-events-none absolute z-10"
            style={{
              left: pos.x,
              top: pos.y,
              transform: "translate(-50%, -110%)",
              transition: "opacity 400ms ease",
              opacity: pos.visible ? 1 : 0,
              willChange: "left, top",
            }}
          >
            <span
              className="text-xl leading-none text-green-700 animate-bounce-slow"
              style={{
                fontWeight: 200,
                WebkitTextStroke: "0.5px rgb(21 128 61)",
              }}
            >
              $
            </span>
          </div>

          <div className="grid grid-cols-1 gap-px overflow-hidden rounded-xl border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-4">
            {benefitItems.map((it, i) => {
              const isActive = i === active;
              return (
                <div
                  key={it.k}
                  ref={(el) => {
                    cardRefs.current[i] = el;
                  }}
                  className={cnLocal(
                    "relative bg-black p-6 transition-all duration-500",
                    isActive ? "bg-zinc-950" : "hover:bg-zinc-950/60"
                  )}
                >
                  <div
                    className={cnLocal(
                      "font-mono text-xs transition-all duration-500",
                      isActive
                        ? "text-green-700 text-[15px] tracking-wider"
                        : "text-zinc-600"
                    )}
                  >
                    {it.k}
                  </div>
                  <h3
                    className={cnLocal(
                      "mt-3 text-base font-semibold transition-colors duration-500",
                      isActive ? "text-white" : "text-zinc-200"
                    )}
                  >
                    {it.h}
                  </h3>
                  <p className="mt-2 text-sm text-zinc-400 leading-relaxed">
                    {it.p}
                  </p>
                  <span
                    className={cnLocal(
                      "absolute left-0 top-0 h-full w-px transition-colors duration-500",
                      isActive ? "bg-green-800/70" : "bg-transparent"
                    )}
                  />
                </div>
              );
            })}
          </div>
          </div>
        </div>
      </div>
    </section>
  );
}

// tiny local cn so we don't pull cn import twice
function cnLocal(...c: Array<string | false | undefined>) {
  return c.filter(Boolean).join(" ");
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
                <pre className="mt-3 overflow-x-auto rounded-md border border-white/5 bg-black px-4 py-3 font-mono text-sm text-zinc-200">
                  <code>
                    <span className="text-zinc-600">$ </span>
                    {s.cmd}
                  </code>
                </pre>
              )}
              {s.note && (
                <p className="mt-3 text-sm text-zinc-500">{s.note}</p>
              )}
            </li>
          ))}
        </ol>

        <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
          <a href="https://github.com/islamborghini/cogni/releases/latest">
            <Button size="lg">
              <Download className="h-4 w-4" />
              Latest release
            </Button>
          </a>
          <a href="https://github.com/islamborghini/cogni#readme">
            <Button variant="outline" size="lg">
              Docs
              <ArrowRight className="h-4 w-4" />
            </Button>
          </a>
          <a
            href="https://github.com/islamborghini/cogni"
            className="inline-flex"
          >
            <Button variant="ghost" size="lg">
              <Github className="h-4 w-4" />
              View source
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
          <div className="h-6 w-6 rounded-md border border-white/15 bg-white/5 grid place-items-center">
            <span className="font-mono text-[11px]">c</span>
          </div>
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
        <Problem />
        <Benefits />
        <DownloadSection />
      </main>
      <Footer />
    </div>
  );
}
