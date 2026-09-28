import { ImageResponse } from "next/og";

export const alt = "Shipcheck: know you're ready before you ship";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const lines: ["pass" | "warn", string][] = [
  ["pass", "Working tree clean"],
  ["pass", "Tests passed"],
  ["pass", "Production build passed"],
  ["pass", "Environment checked"],
  ["warn", "No CI configuration"],
];

// Glyphs are drawn as SVG: the default OG font has no ✓ or ⚠, and
// fetching a fallback font would need network access at build time.
function Tick({ color, size = 22 }: { color: string; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M5 12.5l4.5 4.5L19 7.5" stroke={color} strokeWidth="3" strokeLinecap="square" />
    </svg>
  );
}

function Warn({ size = 22 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M12 3L22 20H2L12 3Z" stroke="#f1c54c" strokeWidth="2.5" strokeLinejoin="round" />
      <path d="M12 10v4.5M12 17v.5" stroke="#f1c54c" strokeWidth="2.5" strokeLinecap="round" />
    </svg>
  );
}

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          background: "#fafaf7",
          padding: 72,
          fontFamily: "sans-serif",
          color: "#0b0b0c",
        }}
      >
        <div style={{ display: "flex", flexDirection: "column", justifyContent: "space-between", width: 560 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 16, fontSize: 34, fontWeight: 700 }}>
            <div
              style={{
                width: 44,
                height: 44,
                background: "#d4f75c",
                border: "2px solid #0b0b0c",
                borderRadius: 6,
                boxShadow: "3px 3px 0 #0b0b0c",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Tick color="#0b0b0c" size={28} />
            </div>
            Shipcheck
          </div>
          <div style={{ display: "flex", flexDirection: "column" }}>
            <div style={{ fontSize: 68, fontWeight: 700, lineHeight: 1.05, letterSpacing: -2 }}>
              Know you&apos;re ready before you ship.
            </div>
            <div style={{ marginTop: 24, fontSize: 28, color: "#55555c", lineHeight: 1.35 }}>
              A minimal open-source CLI that checks your project for common release blockers.
            </div>
          </div>
        </div>
        <div
          style={{
            marginLeft: "auto",
            alignSelf: "center",
            width: 440,
            display: "flex",
            flexDirection: "column",
            background: "#0f1011",
            border: "2px solid #0b0b0c",
            borderRadius: 8,
            boxShadow: "8px 8px 0 #0b0b0c",
            padding: "28px 32px",
            color: "#e8e8e3",
            fontSize: 24,
            fontFamily: "monospace",
          }}
        >
          <div style={{ color: "#8b9099" }}>$ shipcheck</div>
          <div style={{ marginTop: 18, fontWeight: 700 }}>SHIPCHECK</div>
          <div style={{ display: "flex", flexDirection: "column", marginTop: 14, gap: 6 }}>
            {lines.map(([status, text]) => (
              <div key={text} style={{ display: "flex", alignItems: "center", gap: 14 }}>
                {status === "pass" ? <Tick color="#6ad48a" /> : <Warn />}
                {text}
              </div>
            ))}
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 14, marginTop: 18, color: "#6ad48a", fontWeight: 700 }}>
            <Tick color="#6ad48a" />
            Ready to ship
          </div>
        </div>
      </div>
    ),
    size,
  );
}
