import type { Metadata } from "next";

import { Sidebar } from "@/components/sidebar";
import { ThemeProvider } from "@/components/theme-provider";

import "./globals.css";

export const metadata: Metadata = {
  title: "my-agents",
  description: "Self-hosted, always-on AI agents with personas and your own model.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning className="h-full antialiased">
      <body className="flex h-full min-h-screen">
        <ThemeProvider>
          <Sidebar />
          <main className="flex-1 overflow-y-auto">{children}</main>
        </ThemeProvider>
      </body>
    </html>
  );
}
