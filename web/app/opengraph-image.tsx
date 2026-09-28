import { ImageResponse } from "next/og";

export const alt = "Shipcheck: know you're ready before you ship";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const lines: [string, string, string][] = [
  ["✓", "#6ad48a", "Working tree clean"],
  ["✓", "#6ad48a", "Tests passed"],
  ["✓", "#6ad48a", "Production build passed"],
  ["✓", "#6ad48a", "Environment checked"],
  ["⚠", "#f1c54c", "No CI configuration"],
];

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
                fontSize: 28,
              }}
            >
              ✓
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
            {lines.map(([sym, color, text]) => (
              <div key={text} style={{ display: "flex", gap: 12 }}>
                <span style={{ color }}>{sym}</span>
                {text}
              </div>
            ))}
          </div>
          <div style={{ marginTop: 18, color: "#6ad48a", fontWeight: 700 }}>✓ Ready to ship</div>
        </div>
      </div>
    ),
    size,
  );
}
