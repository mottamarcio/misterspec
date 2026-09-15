# Quickstart: Project Documentation Site and README

## 1. Preview the site locally (no build step)

```sh
cd site
python3 -m http.server 8000
# open http://localhost:8000/
```

Any static file server works — `site/` is plain HTML/CSS/JS, nothing
to compile.

## 2. Verify a command's documented example is still accurate

```sh
go build -o /tmp/misterspec ./cmd/misterspec
/tmp/misterspec internal context SPEC-014 --intent implementation --dir <example-project>
# compare the real output against commands.html's own documented example
```

Repeat for any command whose own source changed since the site was
last written (FR-009).

## 3. How publishing works

Every push to `dev` triggers `.github/workflows/deploy-docs.yml`,
which uploads `site/`'s own contents as a Pages artifact and deploys
it — no manual step, no separate hosting account. The published URL
is `https://<owner>.github.io/misterspec/` once the repository's own
Pages settings are switched to "GitHub Actions" as the source (a
one-time, manual repository-settings step outside this feature's own
file changes).

## 4. Adding a new page later

1. Add one entry to `site/assets/nav.js`'s own `NAV` array.
2. Create the new `.html` file, copying an existing page's own
   `<head>` (Tailwind + Iconify CDN script tags, `nav.js` include) and
   `<div id="sidebar">` placeholder.
3. No other file needs to change — every existing page's sidebar
   updates automatically on next load (research.md #2).
