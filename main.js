#!/usr/bin/env node
"use strict";

// Minimal streaming prompt for OpenRouter API.
// Usage:
//   OPENROUTER_API_KEY=... node main.js "Tell me a joke"
//   echo "Explain async/await" | OPENROUTER_API_KEY=... node main.js

const API_URL = "https://openrouter.ai/api/v1/chat/completions";
const MODEL = process.env.OPENROUTER_MODEL || "openai/gpt-4o-mini";
const API_KEY = process.env.OPENROUTER_API_KEY;

if (!API_KEY) {
  console.error("Missing OPENROUTER_API_KEY env var.");
  process.exit(1);
}

async function readStdin() {
  if (process.stdin.isTTY) return "";
  return new Promise((resolve, reject) => {
    let data = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => {
      data += chunk;
    });
    process.stdin.on("end", () => resolve(data.trim()));
    process.stdin.on("error", reject);
  });
}

async function getPrompt() {
  const fromArgs = process.argv.slice(2).join(" ").trim();
  if (fromArgs) return fromArgs;
  const fromStdin = await readStdin();
  return fromStdin;
}

function parseSseLine(line) {
  if (!line.startsWith("data:")) return null;
  const data = line.slice(5).trim();
  if (!data || data === "[DONE]") return null;
  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
}

async function streamPrompt(prompt) {
  const res = await fetch(API_URL, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${API_KEY}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      model: MODEL,
      stream: true,
      messages: [{ role: "user", content: prompt }],
    }),
  });

  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => "");
    throw new Error(`OpenRouter error ${res.status}: ${text}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder("utf-8");
  let buffer = "";

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    let idx;
    while ((idx = buffer.indexOf("\n")) !== -1) {
      const line = buffer.slice(0, idx).trimEnd();
      buffer = buffer.slice(idx + 1);
      const parsed = parseSseLine(line);
      if (!parsed) continue;
      const delta = parsed?.choices?.[0]?.delta?.content;
      if (delta) process.stdout.write(delta);
    }
  }
}

async function main() {
  const prompt = await getPrompt();
  if (!prompt) {
    console.error("No prompt provided. Pass text as args or via stdin.");
    process.exit(1);
  }
  await streamPrompt(prompt);
  process.stdout.write("\n");
}

main().catch((err) => {
  console.error(err?.message || err);
  process.exit(1);
});
