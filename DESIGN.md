# Forge — Design System

> A terminal-first design system inspired by Warp, built for command-line products in 2026.
> Version 1.0 · Dark-mode native · Geist + Geist Mono

---

## Table of contents

1. [Principles](#principles)
2. [File structure](#file-structure)
3. [Tokens](#tokens)
   - [Color](#color)
   - [Typography](#typography)
   - [Spacing](#spacing)
   - [Radii](#radii)
   - [Elevation](#elevation)
   - [Motion](#motion)
4. [Semantic tokens](#semantic-tokens)
5. [Syntax highlighting](#syntax-highlighting)
6. [Components](#components)
7. [Theming & tweaks](#theming--tweaks)
8. [Using the prototype](#using-the-prototype)

---

## Principles

| # | Name | Rule |
|---|------|------|
| 01 | **Quiet by default** | Most pixels are neutral. Color earns attention only when it *is* the message. |
| 02 | **Mono is structure** | `Geist Mono` carries data. `Geist` sans carries narrative. The two never blur. |
| 03 | **AI feels different** | Plasma violet is reserved for agentic surfaces — never compete with user intent. |
| 04 | **Every command is a block** | Status, input, output — addressable, replayable, shareable. The atom of the system. |

---

## File structure

```
tokens.css          — all CSS custom properties (colors, type, spacing, radii, motion)
components.css      — component-level styles (buttons, inputs, blocks, badges, etc.)
ds-cards.jsx        — React: token + component showcase cards for the design canvas
terminal-app.jsx    — React: fully interactive terminal prototype
design-canvas.jsx   — starter: pan/zoom canvas for presenting artboards
tweaks-panel.jsx    — starter: in-page Tweaks panel + protocol
Forge Design System.html — main entry point
```

---

## Tokens

### Color

Forge has five named accent ramps plus a 10-step neutral scale ("Ink").
Each accent has stops 300–700. Always prefer the semantic token over a raw ramp value.

#### Ember (primary accent)

| Token | Value | Use |
|-------|-------|-----|
| `--ember-300` | `#ffb199` | Subtle tint, hover text |
| `--ember-400` | `#ff8a6a` | Hover state on primary |
| `--ember-500` | `#ff6b3d` | **Primary accent** |
| `--ember-600` | `#e85522` | Pressed state |
| `--ember-700` | `#b83d12` | Text on light bg |

#### Plasma (AI / agentic)

| Token | Value | Use |
|-------|-------|-----|
| `--plasma-300` | `#c4b5ff` | AI text on dark |
| `--plasma-400` | `#a78bff` | AI hover |
| `--plasma-500` | `#7c5cff` | **AI primary** |
| `--plasma-600` | `#5e3fe8` | AI pressed |
| `--plasma-700` | `#4527b8` | AI text on light |

#### Status ramps

| Ramp | 400 | 500 | Use |
|------|-----|-----|-----|
| **Mint** | `#4ade94` | `#22c97e` | Success, git-add |
| **Solar** | `#facc4a` | `#eab308` | Warning, cursor |
| **Crimson** | `#f87171` | `#ef4444` | Error, destructive, git-delete |
| **Sky** | `#60a5fa` | `#3b82f6` | Info, links, git-modified |

#### Ink (neutral scale)

| Token | Value | Role |
|-------|-------|------|
| `--ink-0` | `#0a0b10` | Deepest canvas |
| `--ink-1` | `#0f1118` | Primary surface |
| `--ink-2` | `#161922` | Raised surface |
| `--ink-3` | `#1e2230` | Hover state |
| `--ink-4` | `#2a2f40` | Active / strong border |
| `--ink-5` | `#3a4055` | Dividers |
| `--ink-6` | `#5a6178` | Subtle text |
| `--ink-7` | `#8b91a8` | Secondary text |
| `--ink-8` | `#c5c9d6` | Primary text (body) |
| `--ink-9` | `#f0f1f5` | Highest contrast text |

---

### Typography

Two families. Never mix them within a single UI unit.

| Role | Family | Size | Weight | Tracking |
|------|--------|------|--------|----------|
| Display | Geist | 64px | 700 | −0.035em |
| Heading 1 | Geist | 36px | 600 | −0.02em |
| Heading 2 | Geist | 28px | 600 | −0.02em |
| Heading 3 | Geist | 22px | 600 | — |
| Heading 4 | Geist | 18px | 600 | — |
| Body | Geist | 14px | 400 | −0.005em |
| Small | Geist | 12px | 400 | — |
| Mono | Geist Mono | 13px | 400 | 0 |
| Eyebrow / Tag | Geist Mono | 11px | 500 | +0.04em · UPPER |

**CSS variables:**

```css
--font-display: 'Geist', system-ui, sans-serif;
--font-body:    'Geist', system-ui, sans-serif;
--font-mono:    'Geist Mono', ui-monospace, monospace;

--fs-mono-sm: 12px;
--fs-mono:    13px;
--fs-mono-lg: 14px;
```

**Alternate mono fonts** (user-switchable via Tweaks):
- `JetBrains Mono` — wider, clearer at small sizes
- `IBM Plex Mono` — humanist, warmer

---

### Spacing

4px base grid. Use `--space-N` tokens; never hardcode pixel values.

| Token | Value |
|-------|-------|
| `--space-1` | 4px |
| `--space-2` | 8px |
| `--space-3` | 12px |
| `--space-4` | 16px |
| `--space-5` | 20px |
| `--space-6` | 24px |
| `--space-8` | 32px |
| `--space-10` | 40px |
| `--space-12` | 48px |
| `--space-16` | 64px |
| `--space-20` | 80px |

---

### Radii

| Token | Value | Typical use |
|-------|-------|-------------|
| `--r-xs` | 3px | Badges, kbd |
| `--r-sm` | 5px | Tabs, small chips |
| `--r-md` | 8px | Inputs, buttons |
| `--r-lg` | 12px | Command blocks |
| `--r-xl` | 16px | Cards, modals, palettes |
| `--r-2xl` | 20px | Large containers |
| `--r-full` | 9999px | Pills, dots |

> All radius tokens scale proportionally when the user changes **Radius scale** in Tweaks.

---

### Elevation

Three tiers plus accent glow. Prefer the lowest tier that separates the element sufficiently.

| Token | Use |
|-------|-----|
| `--shadow-1` | Buttons, inputs — subtle lift |
| `--shadow-2` | Dropdowns, tooltips — clear separation |
| `--shadow-3` | Modals, command palette — full floating |
| `--shadow-glow-accent` | Focused command block, primary CTA hover |
| `--shadow-glow-ai` | AI suggestion cards |

---

### Motion

| Token | Value | Use |
|-------|-------|-----|
| `--ease-out` | `cubic-bezier(0.16,1,0.3,1)` | Enter transitions |
| `--ease-in-out` | `cubic-bezier(0.65,0,0.35,1)` | State changes |
| `--dur-fast` | 120ms | Hover, focus |
| `--dur-base` | 200ms | Show/hide |
| `--dur-slow` | 320ms | Layouts, overlays |

---

## Semantic tokens

Always use semantic tokens in component code. Raw ramp values (`--ember-500`) are for the token definition layer only.

### Backgrounds

| Token | Dark value | Light value | Use |
|-------|-----------|-------------|-----|
| `--bg-canvas` | ink-0 | ink-1 | Page background |
| `--bg-surface` | ink-1 | ink-0 | Cards, panels |
| `--bg-raised` | ink-2 | ink-2 | Floating elements |
| `--bg-hover` | ink-3 | ink-3 | Hover state |
| `--bg-active` | ink-4 | ink-4 | Active / pressed |

### Foregrounds

| Token | Use |
|-------|-----|
| `--fg-primary` | Headings, strong labels |
| `--fg-secondary` | Body text, secondary content |
| `--fg-muted` | Timestamps, metadata, placeholders |
| `--fg-subtle` | Disabled, decorative |

### Borders

| Token | Dark | Light |
|-------|------|-------|
| `--border-subtle` | rgba(white, 6%) | rgba(black, 6%) |
| `--border-default` | rgba(white, 10%) | rgba(black, 10%) |
| `--border-strong` | rgba(white, 18%) | rgba(black, 18%) |

### Accent (dynamic — changes with theme)

| Token | Role |
|-------|------|
| `--accent` | Primary interactive color |
| `--accent-hover` | Hovered accent |
| `--accent-fg` | Text on accent background |
| `--accent-soft` | Tinted background for selected states |
| `--accent-line` | Border / ring on focused elements |

---

## Syntax highlighting

Used inside command block output areas. Consistent with popular dark-mode themes.

| Token | Dark value | Role |
|-------|-----------|------|
| `--syn-comment` | `#5a6178` | `# comments` |
| `--syn-keyword` | `#ff8a6a` | `git`, `npm`, shell builtins |
| `--syn-string` | `#4ade94` | `"quoted strings"` |
| `--syn-number` | `#facc4a` | `42`, `3.14` |
| `--syn-fn` | `#60a5fa` | function names |
| `--syn-var` | `#c4b5ff` | variable names |
| `--syn-op` | `#c5c9d6` | operators, plain text |
| `--syn-flag` | `#f87171` | `--flags`, `-f` |

CSS utility classes: `.syn-comment`, `.syn-keyword`, `.syn-string`, `.syn-number`, `.syn-fn`, `.syn-var`, `.syn-op`, `.syn-flag`

---

## Components

### Buttons

```html
<button class="btn btn-primary">Primary</button>
<button class="btn btn-secondary">Secondary</button>
<button class="btn btn-ghost">Ghost</button>
<button class="btn btn-danger">Destructive</button>
<button class="btn btn-ai">Ask AI</button>

<!-- Sizes: btn-sm (26px) · default (32px) · btn-lg (40px) -->
```

### Inputs

```html
<!-- Plain -->
<input class="input" placeholder="Value"/>

<!-- With icon/adornment -->
<div class="input-group">
  <svg class="icon">…</svg>
  <input placeholder="Search…"/>
  <span class="kbd">⌘K</span>
</div>
```

### Badges

```html
<span class="badge">default</span>
<span class="badge badge-accent">fast</span>
<span class="badge badge-ok">passed</span>
<span class="badge badge-warn">deprecated</span>
<span class="badge badge-error">failed</span>
<span class="badge badge-info">v1.2</span>
<span class="badge badge-ai">AI</span>
```

### Keyboard shortcut

```html
<span class="kbd">⌘</span><span class="kbd">K</span>
```

### Tabs

```html
<div class="tabs">
  <button class="tab active">~/forge</button>
  <button class="tab">api-server</button>
</div>
```

### Command block

The atomic primitive. Every executed command lives in one.

```html
<div class="cmd-block">
  <div class="cmd-header">
    <div class="status-dot ok"></div>
    <span>~/forge</span> · exit 0 · 1.2s
    <span class="badge badge-ok">success</span>
  </div>
  <div class="cmd-input-line">
    <span class="cmd-prompt">❯</span>
    <span>git status</span>
  </div>
  <div class="cmd-output">…</div>
</div>
```

Status dot classes: `.ok` · `.error` · `.running`  
Add `.focused` to `.cmd-block` for the accent-bordered focused state.

### AI suggestion card

```html
<div class="ai-suggest">
  <div class="ai-suggest-header"><!-- AI label + badge --></div>
  <div><!-- explanation text --></div>
  <div class="ai-suggest-cmd">npm install react@18 react-dom@18</div>
  <div class="ai-suggest-actions">
    <button class="btn btn-ai btn-sm">Run</button>
    <button class="btn btn-ghost btn-sm">Explain more</button>
  </div>
</div>
```

### Modal

```html
<div class="modal-backdrop">
  <div class="modal">
    <div class="modal-header">…</div>
    <div class="modal-body">…</div>
    <div class="modal-footer">
      <button class="btn btn-ghost">Cancel</button>
      <button class="btn btn-danger">Confirm</button>
    </div>
  </div>
</div>
```

### Command palette

Trigger with `⌘K`. Structure:

```html
<div class="cmd-palette">
  <div class="cmd-palette-input">
    <input placeholder="Run a command…"/>
    <span class="kbd">esc</span>
  </div>
  <div class="cmd-palette-list">
    <div class="cmd-palette-section">Workflows</div>
    <div class="cmd-palette-item active">
      <span class="leading">…icon…</span>
      <span class="label">Deploy to production</span>
      <span class="meta">forge deploy</span>
    </div>
  </div>
</div>
```

---

## Theming & tweaks

The system supports full runtime theming via CSS custom properties.

### Dark / Light

```html
<!-- Dark (default — no attribute needed) -->
<html>

<!-- Light -->
<html data-theme="light">
```

### Density

```html
<html data-density="compact">   <!-- 0.85× spacing multiplier -->
<html data-density="default">   <!-- 1.0× -->
<html data-density="cozy">      <!-- 1.15× -->
```

### Accent swapping

Override the `--ember-*` ramp and the semantic `--accent*` tokens to change the primary color. Five presets ship out of the box: **Ember** (coral), **Plasma** (violet), **Mint** (green), **Sky** (blue), **Solar** (yellow).

### In-page Tweaks panel

Open with the **Tweaks** toggle in the toolbar. Controls:
- Dark / Light mode
- Primary accent (5 options)
- Mono font (Geist Mono · JetBrains Mono · IBM Plex Mono)
- Density (Compact · Default · Cozy)
- Radius scale (0.25× → 2×)

---

## Using the prototype

Open `Forge Design System.html`. The design canvas has four sections:

| Section | Contents |
|---------|----------|
| **Overview** | Principles card |
| **The product** | Live interactive terminal |
| **Foundation** | Color palette · Typography · Spacing/radii/elevation |
| **Components** | Buttons/inputs/badges · Command blocks · Overlays/AI/modal |

### Terminal commands (interactive prototype)

| Command | What it does |
|---------|-------------|
| `ls` | List directory |
| `git status` / `git log` / `git branch` | Git output |
| `npm install` / `npm test` / `npm run dev` | npm output |
| `forge deploy` / `forge status` / `forge logs` | Deploy flow |
| `ai <question>` | Invokes AI suggestion |
| `clear` | Clears all blocks |
| `help` | Lists all commands |
| `⌘K` | Opens command palette |

---

*Forge is a design artifact — the commands are simulated. The component tokens and CSS are production-ready.*
