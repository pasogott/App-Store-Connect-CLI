# Build and package iOS apps on Linux

ASC's portable workflow uses an installed Swift/Apple SDK toolchain, `xtool`,
and `rcodesign`. The tested xtool backend is
[joshuaswarren/xtool at f0a1f90](https://github.com/joshuaswarren/xtool/tree/f0a1f90efdbb0dc023e276ff529da92618da7a87);
it must replace its app output directory on a successful build so ASC can reject
stale output. It does not install these tools or provide an iOS simulator
runtime. The initial backend supports one thin arm64 iOS app plus resources;
nested executable code, extensions, embedded frameworks, watch apps, App Clips,
and app symlinks are rejected.

## User-supplied tooling and SDK licensing

ASC provides a local integration with tools installed by the user. It does not
bundle, download, or redistribute Xcode, Apple SDKs, or Swift toolchains, and it
does not operate a hosted build service. Users must obtain and configure their
own toolchain and signing materials and are responsible for reviewing and
complying with the agreements applicable to their account and build environment.

This follows the tooling boundary described by xtool's author,
[Kabir Oberai](https://forums.swift.org/t/xtool-cross-platform-xcode-replacement-build-ios-apps-on-linux-and-more/79803/3):
xtool does not distribute Apple's SDKs or toolchains, asks users to supply Xcode,
and leaves license compliance to users. He also suggests using macOS CI for the
final App Store build when concerned about a Linux build environment. Operators
can use a macOS/Xcode workflow for their final distribution build.

Providing an integration does not establish permission to use an SDK in every
environment. Apple's standard
[Developer Program agreement](https://developer.apple.com/support/terms/apple-developer-program-license-agreement/)
§2.6 restricts Apple SDK use on non-Apple computers and enabling others to do so.
The [Xcode/SDK agreement](https://www.apple.com/legal/sla/docs/xcode.pdf) authorizes
execution on Apple hardware running macOS. No applicable exception or written
Apple permission has been verified. User-supplied tooling does not waive these
terms or establish Apple approval of this workflow. Technical verification does
not establish licensing compliance or App Store acceptance.

For an Xcode project, first prepare a Swift package using the converter from
[omarchy-apple-dev](https://github.com/joshuaswarren/omarchy-apple-dev):

```sh
python3 /opt/omarchy-apple-dev/tools/xcodeproj2xtool.py App.xcodeproj --target App
```

The converter creates `omarchy-xtool/`. Inspect its generated Package.swift,
xtool.yml, Info.plist, dependencies, and resource handling. Xcode schemes,
workspaces, build settings, and arbitrary scripts are not interchangeable with
Swift package targets. Keep the Apple SDK/toolchain version pinned.

Build a release device app:

```sh
asc builds compile --package-path ./omarchy-xtool --product App \
  --platform device --configuration release --app-path ./artifacts/App.app
```

Package it with an existing local identity and provisioning profile:

```sh
asc builds package --app-path ./artifacts/App.app --ipa-path ./artifacts/App.ipa \
  --identity ./signing/App.p12 --identity-password-file ./secrets/password \
  --provisioning-profile ./signing/App.mobileprovision --entitlements ./signing/entitlements.plist
```

The profile must be signed by Apple's trusted provisioning signer, unexpired,
and match the exact app bundle ID and supplied signing certificate. Requested
entitlements must be authorized by that profile. Optional capabilities are not
automatically copied from the profile or an existing code signature. This
command packages fresh build output; use `asc signing resign` on macOS for its
existing complete IPA re-signing contract.

For a local pipeline test without an Apple identity:

```sh
asc builds package --app-path ./artifacts/App.app --ipa-path ./artifacts/Test.ipa --ad-hoc
```

Ad hoc test output is not accepted by Apple and cannot be installed on an
ordinary physical iPhone. A simulator app is a separate build target:

```sh
asc builds compile --package-path ./omarchy-xtool --product App \
  --platform simulator --configuration debug --app-path ./artifacts/Simulator.app
```

Simulator output gets corrected platform metadata and a Linux-created ad hoc
signature. A Mac iOS simulator host is still needed to install and run it. A
simulator executable cannot be packaged as a device IPA.

Compiler and signer diagnostics go to stderr. JSON (or explicit table/Markdown)
receipts go to stdout. Outputs are create-only; source apps are copied into
private staging before signing. Inspect a reported incomplete app publication
before retrying. Existing Xcode archive/export/resign behavior is unchanged.

`signatureVerified: false` is intentional: rcodesign completion is not Apple's
complete bundle signature verification. `profileValidated: true` reports the
specific profile/identity/entitlement authorization checks, not App Store
acceptance. Apple metadata requirements, assets, capabilities, SDK stamping,
debug symbols, and project compatibility still need validation before shipping.

Upload a genuinely signed, prepared IPA as a separate authorized step:

```sh
asc builds upload --app APP_ID --ipa ./artifacts/App.ipa --wait
```

Neither portable command discovers Apple credentials, creates certificates or
profiles, invokes upstream ship.sh, or uploads anything. rcodesign configuration
is disabled and timestamp requests are disabled for local signing.
