<p align="center">
  <img width="192px" src="https://github.com/emacs-mirror/emacs/raw/emacs-27.2/etc/images/icons/hicolor/scalable/apps/emacs.svg" alt="Logo">
</p>

<h1 align="center">
  Emacs Builds
</h1>

<p align="center">
  <a href="https://github.com/jimeh/emacs-builds/releases">
    <img src="https://img.shields.io/github/v/tag/jimeh/emacs-builds?label=nightly" alt="GitHub tag (latest SemVer)">
  </a>
  <a href="https://github.com/jimeh/emacs-builds/issues">
    <img src="https://img.shields.io/github/issues-raw/jimeh/emacs-builds.svg?style=flat&logo=github&logoColor=white"
alt="GitHub issues">
  </a>
  <a href="https://github.com/jimeh/emacs-builds/pulls">
    <img src="https://img.shields.io/github/issues-pr-raw/jimeh/emacs-builds.svg?style=flat&logo=github&logoColor=white" alt="GitHub pull requests">
  </a>
</p>

<p align="center">
  <strong>
    Nightly binary builds of Emacs for macOS as a self-contained Emacs.app,
    with native-compilation.
  </strong>
</p>

## Features

- Self-contained Emacs.app application bundle, with no external dependencies,
  built with the [jimeh/build-emacs-for-macos][] build script.
- Native-compilation is enabled in nightly builds from the `master` branch of
  Emacs, and should just work without the need to install GCC, libgccjit, or any
  other dependencies.
- Includes the [fix-window-role][] and [system-appearance][] patches from the
  excellent [emacs-plus][] project.
- Emacs source is fetched from the [emacs-mirror/emacs][] GitHub repository.
- Build creation is transparent and public through the use of GitHub Actions,
  allowing anyone to inspect git commit SHAs, full source code, and exact
  commands used to produce a build. This is especially important right now as
  builds are not yet signed and notarized.

[jimeh/build-emacs-for-macos]: https://github.com/jimeh/build-emacs-for-macos
[fix-window-role]:
  https://github.com/d12frosted/homebrew-emacs-plus/blob/master/patches/emacs-28/fix-window-role.patch
[system-appearance]:
  https://github.com/d12frosted/homebrew-emacs-plus/blob/master/patches/emacs-28/system-appearance.patch
[emacs-plus]: https://github.com/d12frosted/homebrew-emacs-plus
[emacs-mirror/emacs]: https://github.com/emacs-mirror/emacs

## System Requirements

- Intel-based Mac running macOS 10.15.x or later.

## Downloads

See the [Releases][] page to download latest builds.

Nightly builds of Emacs are for the part just fine, but if you don't like living
too close to the edge, see issue [#7 Known Good Nightly Builds][7] for a list of
recent nightly builds which have been actively used by a living being without
any issues.

[releases]: https://github.com/jimeh/emacs-builds/releases
[7]: https://github.com/jimeh/emacs-builds/issues/7

## Build Process

Building Emacs is done using the [jimeh/build-emacs-for-macos][] build script,
executed within a GitHub Actions [workflow][]. This is why macOS 10.15.x or
later is required, as it's the oldest version of macOS available in GitHub
Actions.

[workflow]:
  https://github.com/jimeh/emacs-builds/blob/main/.github/workflows/build.yml

Full history for all builds is available on GitHub Actions [here][actions].

[jimeh/build-emacs-for-macos]: https://github.com/jimeh/build-emacs-for-macos
[actions]: https://github.com/jimeh/emacs-builds/actions

Nightly builds are scheduled for 0:00 UTC every night, based on the latest
commit from the `master` branch of the [emacs-mirror/emacs][] repository. This
means a nightly build will only be produced if there have been new commits since
the last nightly build.

[emacs-mirror/emacs]: https://github.com/emacs-mirror/emacs

## Untrusted Application

Currently builds are not signed or notarized, meaning macOS cannot verify
Emacs.app came from a trusted developer, and by default you are not given an
option to trust the app and open it.

Simplest way around this is to right-click (or control-click) on the Emacs app
in Finder and select "Open". You will then be given the same warning as before,
but with a "Open" button now available to trust and open the app. After that you
can open the application like normal without any warnings.

## Issues / To-Do

Please see [Issues][] for details of things to come, or to report issues.

[issues]: https://github.com/jimeh/emacs-builds/issues
