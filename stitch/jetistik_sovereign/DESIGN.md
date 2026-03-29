# Design System Specification: The Sovereign Ledger

## 1. Overview & Creative North Star
**Creative North Star: "The Digital Architect"**
This design system moves away from the clinical, cold nature of traditional government portals. Instead, it adopts the persona of a "Digital Architect"—authoritative, precise, and sophisticated. We are not just building a form-filler; we are creating a digital ledger of merit. 

The aesthetic is defined by **High-End Editorial Minimalism**. We challenge the rigid, "boxed-in" grid by utilizing intentional asymmetry, expansive breathing room, and a hierarchy driven by tonal depth rather than structural lines. This system feels "government-adjacent" through its stability but "modern-tech" through its fluid, glass-like surfaces and refined typography.

---

## 2. Colors & Surface Philosophy
The palette is rooted in an institutional Blue (`primary`), but elevated through a nuanced scale of cool grays and architectural whites.

### The "No-Line" Rule
**Lines are a failure of hierarchy.** Designers are prohibited from using 1px solid borders to section off content. Boundaries must be defined solely through background color shifts. For example, a `surface-container-low` section should sit directly on a `surface` background to create a logical break without visual clutter.

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers—like stacked sheets of fine vellum.
*   **Base Layer:** `surface` (#f7f9fb)
*   **Secondary Content Area:** `surface-container-low` (#f2f4f6)
*   **Interactive Cards/Modules:** `surface-container-lowest` (#ffffff)
*   **Elevated Overlays:** `surface-container-high` (#e6e8ea)

### The "Glass & Gradient" Rule
To avoid a "flat" template look, use Glassmorphism for floating elements (e.g., navigation bars, modal headers). Use a semi-transparent `surface` color with a `backdrop-blur` of 12px-20px. 
*   **Signature Textures:** For primary CTAs and Hero sections, use a subtle linear gradient from `primary` (#004ac6) to `primary_container` (#2563eb) at a 135° angle. This adds a "soul" to the interface that flat hex codes cannot replicate.

---

## 3. Typography
The system uses a pairing of **Manrope** (Display/Headlines) and **Inter** (Body/Labels) to balance institutional authority with technical readability.

*   **Display (Manrope):** Large, bold, and expressive. Use `display-lg` (3.5rem) for hero statements to create an editorial feel.
*   **Headlines (Manrope):** Use `headline-sm` (1.5rem) for section headers. The wider tracking of Manrope provides a "premium" breathability.
*   **Body (Inter):** All functional data, Kazakh/Russian bilingual labels, and certificate details must use `body-md` (0.875rem) or `body-lg` (1rem). 
*   **Bilingual Handling:** Kazakh and Russian labels should be handled with a "Primary/Subtle" pairing. The primary language uses `on_surface`, while the secondary translation uses `on_surface_variant` at `label-md` size, positioned directly below or beside the primary text.

---

## 4. Elevation & Depth
We achieve hierarchy through **Tonal Layering** rather than shadows.

*   **The Layering Principle:** Place a `surface-container-lowest` card on a `surface-container-low` section to create a soft, natural lift. This mimics the way light hits physical paper.
*   **Ambient Shadows:** If an element must "float" (like a certificate preview), use a shadow with a 40px blur and 4% opacity, tinted with the `primary` hue rather than pure black.
*   **The "Ghost Border" Fallback:** If accessibility requires a container edge, use a "Ghost Border": the `outline-variant` token at 15% opacity. Never use 100% opaque borders.
*   **Depth through Blur:** Use `backdrop-filter: blur(10px)` on all modal overlays to keep the user grounded in the "Sovereign Ledger" environment while focusing on the task at hand.

---

## 5. Components

### Professional Forms & Inputs
*   **Input Fields:** Use `surface-container-lowest` as the field background. Instead of a 4-sided border, use a 2px bottom-accent of `outline-variant`. Upon focus, transition the bottom-accent to `primary`.
*   **Validation States:** Error states must use `error` (#ba1a1a) text with a `error_container` background wash. Avoid "scary" high-contrast red boxes; keep it sophisticated.

### Status Badges (The "Validator" Style)
*   **VALID:** Use `on_primary_container` text on a `primary_fixed` background.
*   **REVOKED/INVALID:** Use `on_error_container` text on `error_container`.
*   **Styling:** Pills must use `rounded-full` (9999px) with `label-md` uppercase typography and 0.05em letter spacing.

### Data Tables & Certificate Lists
*   **Forbid Dividers:** Do not use horizontal lines between rows. Use the Spacing Scale `4` (1rem) to create clear air between items.
*   **Alternating Tones:** Use a subtle shift from `surface` to `surface-container-low` for zebra-striping if data density is high.

### Buttons
*   **Primary:** Linear gradient (`primary` to `primary_container`), `rounded-md` (0.375rem), with a soft `primary` ambient shadow.
*   **Tertiary (Ghost):** No background, `on_surface` text. Use for secondary actions like "Download CSV" or "View Metadata."

### QR Code Integration
*   QR codes should never sit on a raw white background. Place them on a `surface-container-lowest` card with a `1.5` (0.375rem) padding and a very subtle `primary` "Ghost Border" to frame the code as a premium security feature.

---

## 6. Do's and Don'ts

### Do:
*   **Do** use asymmetrical margins (e.g., a wider left margin for titles) to create an editorial, high-end look.
*   **Do** prioritize whitespace. If a layout feels "busy," increase the spacing between containers using the `12` (3rem) or `16` (4rem) tokens.
*   **Do** use `surface-tint` sparingly for small iconography to draw the eye to interactive zones.

### Don't:
*   **Don't** use pure black (#000000) for text. Use `on_surface` (#191c1e) to maintain a soft, premium contrast.
*   **Don't** use traditional "Drop Shadows" from 2010. If it doesn't look like ambient natural light, remove it.
*   **Don't** crowd the bilingual text. If Kazakh and Russian are both present, ensure the line-height is at least 1.6x to prevent the scripts from visually merging.
*   **Don't** use sharp corners. Stick to the `md` (0.375rem) or `lg` (0.5rem) roundedness to keep the interface feeling approachable and modern.