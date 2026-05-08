import { cn } from "@/lib/utils";

type Tok = { t: string; c?: string };

const K = "text-violet-300";       // keyword-ish (tool, result)
const S = "text-emerald-300";      // strings
const P = "text-zinc-500";         // punctuation
const N = "text-amber-200";        // numbers / values
const C = "text-zinc-500 italic";  // comments
const F = "text-sky-300";          // identifiers / file paths
const W = "text-white";            // emphasis
const G = "text-zinc-400";         // dim text

const lines: Tok[][] = [
  [{ t: "$ ", c: P }, { t: "claude ", c: G }, { t: "--mcp-config ", c: G }, { t: "cogni.json", c: F }],
  [],
  [{ t: "// agent calls cogni instead of grep + read", c: C }],
  [
    { t: "tool", c: K },
    { t: "(", c: P },
    { t: '"mcp__cogni__repo_overview"', c: S },
    { t: ")", c: P },
  ],
  [
    { t: "  ", c: P },
    { t: "→ ", c: P },
    { t: "23", c: N },
    { t: " packages, ", c: G },
    { t: "412", c: N },
    { t: " symbols", c: G },
  ],
  [],
  [
    { t: "tool", c: K },
    { t: "(", c: P },
    { t: '"mcp__cogni__symbol_search"', c: S },
    { t: ", ", c: P },
    { t: '"AsyncHTTPTransport"', c: S },
    { t: ")", c: P },
  ],
  [
    { t: "  ", c: P },
    { t: "→ ", c: P },
    { t: "_transports/default.py", c: F },
    { t: ":", c: P },
    { t: "42", c: N },
  ],
  [],
  [
    { t: "tool", c: K },
    { t: "(", c: P },
    { t: '"mcp__cogni__symbol_source"', c: S },
    { t: ", ", c: P },
    { t: '"AsyncHTTPTransport.handle_async_request"', c: S },
    { t: ")", c: P },
  ],
  [
    { t: "  ", c: P },
    { t: "→ ", c: P },
    { t: "32", c: N },
    { t: " lines (vs ", c: G },
    { t: "412", c: N },
    { t: "-line file)", c: G },
  ],
  [],
  [
    { t: "result", c: K },
    { t: ": ", c: P },
    { t: "25%+", c: W },
    { t: " fewer tokens", c: G },
  ],
];

export function CodeWindow({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "relative w-full overflow-hidden rounded-xl border border-white/10 bg-[#050505] shadow-2xl",
        className
      )}
    >
      <div className="flex items-center gap-2 border-b border-white/10 bg-black/60 px-4 py-3">
        <span className="h-3 w-3 rounded-full bg-zinc-700" />
        <span className="h-3 w-3 rounded-full bg-zinc-700" />
        <span className="h-3 w-3 rounded-full bg-zinc-700" />
        <span className="ml-3 text-xs text-zinc-500 font-mono">
          ~/code/httpx · cogni session
        </span>
      </div>
      <pre className="font-mono text-[13px] leading-[1.7] p-5 overflow-x-auto">
        <code>
          {lines.map((line, i) => (
            <div key={i} className="grid grid-cols-[2.25rem_1fr] min-h-[1.7em]">
              <span className="select-none text-right pr-4 text-zinc-700">
                {i + 1}
              </span>
              <span>
                {line.length === 0 ? (
                  <span>&nbsp;</span>
                ) : (
                  line.map((tok, j) => (
                    <span key={j} className={tok.c}>
                      {tok.t}
                    </span>
                  ))
                )}
              </span>
            </div>
          ))}
        </code>
      </pre>
    </div>
  );
}
