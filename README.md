# nuguard-canary

> **INTENTIONALLY VULNERABLE TEST CODE. DO NOT USE, DEPLOY OR COPY.**
>
> This repository exists only to check that an AI security scanner (NuGuard,
> through vulnerabilityAI) finds and reports known AI vulnerabilities. Every
> flaw in it is deliberate and labelled. It contains no real credentials:
> the only key-like string is an obviously fake placeholder.

A tiny "support assistant" written in Go (standard library only), plus one
small Python helper. It never talks to a real model: nothing here is run
against a live API.

## Planted findings (expected)

Each planted flaw is in its own function, marked `// CANARY:` (Go) or
`# CANARY:` (Python), with the rule it is meant to trigger.

| # | File | Function | Expected rule | Kind |
|---|---|---|---|---|
| 1 | `assistant/vulnerable.go` | `SummarizeTicket` | `nuguard-go-llm-prompt-injection-sprintf` | prompt injection (LLM01) |
| 2 | `assistant/vulnerable.go` | `TranslateReply` | `nuguard-go-llm-prompt-injection-sprintf` | prompt injection through a hand-written provider call |
| 3 | `assistant/vulnerable.go` | `NewHardcodedClient` | `nuguard-go-hardcoded-api-key` | hard-coded AI API key (fake) |
| 4 | `assistant/vulnerable.go` | `ClassifyNoTimeout` | `nuguard-go-llm-missing-context-timeout` | LLM call without a deadline (LLM04). **Fixed** in the second commit, as a lifecycle test: it should be verified as fixed. |
| 5 | `assistant/vulnerable.go` | `AskOverInsecureTLS` | `nuguard-go-llm-insecure-tls` | TLS verification disabled for LLM traffic |
| 6 | `tools/agent_tools.py` | `ask` | `nuguard-llm-prompt-injection-fstring` | prompt injection (Python) |
| 7 | `tools/agent_tools.py` | `run_generated_code` | `nuguard-eval-on-llm-output` (×2: `eval`, `exec`) | code execution on model output |

## Safe counterparts (must NOT be reported)

`assistant/safe.go` and the `safe_*` functions in `tools/agent_tools.py` do
the same jobs correctly: the key comes from the environment, every call has a
timeout, TLS is left at its defaults, user text is never formatted into the
prompt, and model output is never executed. A finding in any of them is a
false positive.

Other scanners (gosec, gitleaks and so on) may also report some of the planted
flaws, such as `InsecureSkipVerify` or the fake key. Those are application
findings, reported alongside NuGuard's AI findings.

## After the test

Archive this repository on GitHub, and unregister it from vulnerabilityAI
(`./vfctl.sh repo remove pobtampal/nuguard-canary`).
