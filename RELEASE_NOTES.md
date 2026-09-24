# Magus v0.5.1

A maintenance release that makes Homebrew setup more robust on the Mac and
adds continuous checks to the project. There are no catalogue or menu changes.
The manifest schema remains `0.4.0`.

## Fixes

- **Bounded Homebrew setup.** The official Homebrew installer launched from
  Review & install now runs under a one-hour limit and is stopped whenever
  Magus ends the setup or quits, so an abandoned installer can no longer
  outlive the session that started it. Its temporary script and install lock
  are still released on every exit path.
- **Opening Raycast links can no longer hang.** Opening a Raycast manual or
  extension page is limited to ten seconds.

## Project

- Every pull request and push to `main` now runs formatting, `go vet`, and
  race-enabled Go tests on Linux and macOS, plus the website tests, type
  check, and a check that the embedded command catalogue matches the site.
- The website and package metadata now advertise the current release, and the
  README reflects that the Mac alpha is available.

## Upgrade

Choose **Update Magus** from the Mac menu, or run:

```sh
curl -fsSL https://magus.sh/install | sh
```

Existing manifests and profiles continue to work without a schema change.
