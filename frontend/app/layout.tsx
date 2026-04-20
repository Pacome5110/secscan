import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SecScan — Web Security Scanner",
  description:
    "Enter a URL and scan it for open ports, security headers, TLS issues, XSS, SQLi, and known CVEs — all in parallel.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="tr">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap"
          rel="stylesheet"
        />
      </head>
      <body>{children}</body>
    </html>
  );
}
