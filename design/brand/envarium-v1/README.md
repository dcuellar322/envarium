# Envarium visual identity v1

Created October 15, 2026. This is a reviewable brand asset package for the Phase 20 identity slice. It does not change the installed application or public website identity.

## Design

A rounded E brings three environment rails into one spine. A square cutout suggests a terminal cursor. The custom geometric lowercase wordmark shares its rounded geometry. Colors come from the canonical Envarium mockup: signal blue `#78a6ff`, violet `#9e7bff`, canvas `#0a0d12`, panel `#11161e`, and foreground `#edf3fb`.

## Files

| Directory | Contents |
| --- | --- |
| `logos/` | Horizontal lockups for dark and light backgrounds, plus solid white and black; SVG and transparent 2440 × 576 PNG |
| `marks/` | Standalone gradient, blue, white, black, and simplified small-size SVG symbols |
| `app/` | App tile SVG, 64–1024 px PNGs, Windows ICO, and macOS ICNS |
| `web/` | SVG and ICO favicons; 16, 32, 48, 180, 192, and 512 px PNG exports |
| `tray/` | Transparent black and white template PNGs at 16, 22, and 32 px |
| `interface/` | 12 standalone SVG interface icons on a 24 px grid |
| `references/` | Original generated concept and exact generation prompt |

`preview.png` and `preview.svg` show the complete system. The preview tagline is presentation copy and is not baked into the logo files.

## Usage

- Prefer `logos/envarium-on-dark.svg` on dark surfaces and `logos/envarium-on-light.svg` on light surfaces. Logo artwork is entirely paths, with no font or network dependencies.
- Reserve at least one rail thickness around the visible mark. Keep the full horizontal logo at least 160 CSS px wide; use the symbol below that width.
- Use `marks/envarium-small.svg` at 16–24 CSS px. Its cutout is removed for legibility. Tray exports use this same simplified shape.
- Preserve the mark's proportions and transparent padding. Use the monochrome variants where gradients are unsuitable.
- Interface icons use `currentColor`, a 1.8-unit stroke, and rounded caps and joins. Inline the SVG to inherit CSS color. When loading via an image element, select an explicit color in a copied asset if needed.
- Standalone SVGs include accessible names. For decorative inline use, set `aria-hidden="true"` and remove their standalone labeling. For inline logos, give each embedded title a unique ID. Label icon-only buttons by their action.
- PNG/ICO/ICNS files are exports; the SVG files are the editable production artwork. Platform containers were structurally checked, but native packaging and installed OS appearance were not tested.

## Creation and review

The built-in image generation tool produced the initial concept using the exact prompt in `references/prompt.txt`. The final mark was reconstructed as clean SVG geometry, with an original path-based wordmark and code-native interface icons. The generated raster is reference material; use the SVG sources and their clean exports in production.

The package has been visually inspected as a rendered contact sheet. XML, dimensions, transparency, and icon-container structure were checked locally. Applying this identity across the application, desktop packaging, and public site remains a separate integration step after human review.

These assets follow the repository's Apache-2.0 license and existing product attribution guidance.
