# Product screenshots

Sanitized, deterministic captures of the real Vue application with the integrated
Envarium branding. The October 17, 2026 captures come from the macOS Chromium
visual baselines at 1440 × 1050. They contain fixture projects and paths, without
user repositories or credentials.

After changing product UI, run and review the app's visual suite with
`--update-snapshots=all`. Run `pnpm --dir site screenshots --platform=darwin` (or
`linux` for reviewed Linux captures), then rebuild and visually check the site.
The site owns these publication assets so a build on another OS cannot silently
switch the images' typography or retain a former brand.
