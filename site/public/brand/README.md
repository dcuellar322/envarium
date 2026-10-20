# Envarium brand assets

Use these logos, icons, and colors to identify Envarium in articles,
integrations, and project documentation. Clearly identify unofficial builds
and avoid implying endorsement or official distribution.

## Choose an asset

- `logos/`: horizontal logos for dark or light backgrounds and solid white or
  black variants, in SVG and transparent PNG.
- `marks/`: standalone symbols. Use the simplified small mark at 16–24 CSS px.
- `app/`: app-tile SVG, PNG, ICO, and ICNS artwork. These are icons, not installers.
- `web/`: SVG/ICO favicons and PNG web icons.
- `tray/`: black and white template PNGs at 16, 22, and 32 px.
- `interface/`: SVG interface icons; `icons.ts` lists their names.
- `palette.json`: color values; `preview.png` and `preview.svg`: visual overview.
- `social.svg`: social-card artwork.

## Use the assets

Keep the full horizontal logo at least 160 CSS px wide. Below that width, use
the standalone mark. Leave at least one rail thickness of clear space around it.
Preserve proportions and transparent padding. Choose a supplied gradient or
monochrome version that contrasts with its background.

SVG artwork is editable and uses paths, with no font or network dependencies.
Interface icons use a 24 px grid, 1.8-unit strokes, rounded caps, and currentColor.
Inline an icon to inherit CSS color. Label icon-only controls by their action;
mark decorative icons aria-hidden. Use unique title IDs when embedding logos.

The core gradient runs from signal blue #78a6ff to violet #9e7bff. Check text
contrast against the actual background; these decorative colors are not a
contrast guarantee for every text size or surface.

## License and attribution

These assets follow the repository’s Apache-2.0 license. Keep the accompanying
LICENSE and NOTICE when redistributing the kit. Section 6 of the license covers
trademark use. Clearly distinguish forks and unofficial distributions.
The social card’s supporting letterforms are outlined from Liberation Sans
(Google/Red Hat, SIL Open Font License 1.1).

Windows ICO artwork is included for reuse. Windows and WSL2 product support is
coming soon; artwork availability does not indicate installer availability.

## Contributing asset changes

Edit the SVG sources, palette.json, or icons.ts in the repository. Run
pnpm brand:sync to regenerate exports and the downloadable kit, then run
pnpm brand:check. Include the generated files with the source change.
