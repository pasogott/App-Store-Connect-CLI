# Portable iOS build and package workflow

Add `asc builds compile` and `asc builds package`, using explicitly installed xtool
and rcodesign. Preserve existing Xcode archive/export and fully verified
macOS re-signing contracts. xtool consumes a Swift package, not an Xcode scheme;
its converted product is not an Xcode archive.

Build accepts a prepared `--package-path`, `--product`, exact new `--app-path`,
`--platform device|simulator`, and `--configuration debug|release`. The installed
toolchain owns Swift/Apple SDK setup. The existing upstream converter can prepare
Rork Xcode projects. Compiler output goes to stderr; JSON/table/Markdown receipts
go to stdout. Simulator metadata is corrected and the app signed ad hoc.

Package accepts one built device app, an exact new IPA destination, and either
explicit `--ad-hoc` test signing or local PKCS#12 identity and Apple provisioning
profile. Real signing validates Apple CMS profile trust, expiry, exact bundle ID,
certificate/profile binding, and requested entitlement authorization using
existing portable validators. Reject nested code until it has an explicit signing
plan. Do not provision resources, discover credentials, upload, or run upstream
ship.sh (which can implicitly provision signing assets).

rcodesign's verification does not provide Apple's complete bundle verification.
The receipt reports that boundary; real Apple processing remains a separate
`asc builds upload --wait` step. Ad hoc IPAs are local pipeline tests and cannot
be published to Apple. Preserve create-only outputs, rooted reads/writes,
bounded copies, private key staging, cancellation, and sanitized child env.

RED: CLI rejects unsupported platform and missing/conflicting signing selection.
GREEN: real child execution, failure propagation, safe publication, hostile
paths, strict signing validation and output renderer tests. E2B smoke uses the
saved experimental Linux SDK/toolchain image and an ad hoc fixture without Apple
credentials. Existing Xcode/resign tests remain unchanged.

## Compiler output freshness contract

This backend requires xtool to publish a fresh product directory. The tested
xtool fork is `joshuaswarren/xtool` at
`f0a1f90efdbb0dc023e276ff529da92618da7a87`.
Its [Packer.swift](https://github.com/joshuaswarren/xtool/blob/f0a1f90efdbb0dc023e276ff529da92618da7a87/Sources/PackLib/Packer.swift#L123-L125)
removes the prior `xtool/<product>.app` and persists a separately staged bundle
at that path after packing. Swift's compilation cache is independent of this
bundle publication. The pinned old directory handle prevents inode reuse from
being mistaken for the prior output.

The initial review suggested normal incremental builds modify this directory in
place. Source inspection and E2B verification establish otherwise for the
supported backend: a release build against the prebuilt image, a second cached
release build, and restored-source recovery all succeeded with existing xtool
output. The unchanged-directory rejection is deliberate: a successful no-op or
build of another configured product must not publish an old app. Other xtool
versions that publish in place need a separately tested freshness contract
before support is claimed.

## Generated reference verification

The review suggested missing subcommands in `docs/COMMANDS.md`. The current
`scripts/generate-command-docs.py` reads only root `asc --help` and generates
command families, not nested commands. `make generate-command-docs` was rerun
and produced an identical file; `make check-docs` verifies synchronization.
The new subcommands are documented in `docs/linux-ios.md`, embedded ASC guide,
and live `asc builds --help`. Expanding the generator's scope is unrelated to
this change; manually changing its output would fail the synchronization gate.

## Empty entitlement override

The hard audit reproduced a signed-input bug on E2B with rcodesign 0.29.0:
both an omitted `--entitlements` and an explicitly empty plist preserved the
input signature's optional entitlement. rcodesign imports previous claims when
no explicit entitlement setting is provided. Package now always supplies a
valid XML dictionary, including an empty dictionary to clear old claims. The
shared macOS marshal helper remains unchanged; its empty-map result is no bytes,
which cannot be passed to rcodesign as an XML file. Regression tests cover both
empty selection forms; live Linux verification inspects actual output signature
entitlements rather than inferring them from signer arguments.
