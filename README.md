<div align="center">
	<h1 align="center">🧀 Parm 🧀</h1>
	<h4 align="center">Install any program from your terminal.</h4>
	<a href="https://www.gnu.org/licenses/gpl-3.0"><img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License: GPL v3"></a>
	<img alt="GitHub Release" src="https://img.shields.io/github/v/release/yhoundz/parm">
	<a href="https://github.com/yhoundz/parm/actions/workflows/ci.yml"><img src="https://github.com/yhoundz/parm/actions/workflows/ci.yml/badge.svg?branch=master" alt="ci"></a>
	<a href="https://github.com/yhoundz/parm/actions/workflows/release.yml"><img src="https://github.com/yhoundz/parm/actions/workflows/release.yml/badge.svg" alt="release"></a>
	<br>
</div>

> [!IMPORTANT]
> Parm is currently in a pre-release state. Expect breaking changes and bugs.

**Table of Contents:**
1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
    - [What is Parm?](#what-is-parm)
    - [Disclaimers](#disclaimers)
3. [Pre-requisites](#pre-requisites)
4. [Installation](#install)
5. [GitHub Personal Access Token](#adding-a-github-personal-access-token)
6. [Usage/Documentation](#usage)
7. [Contributing](#contributing)
8. [Adding Packages to Parm](#adding-packages-to-parm)
9. [Acknowledgements](#Acknowledgements)

## Introduction

### What is Parm?
Parm is a cross-platform binary installer with a package manager-esque workflow. It allows you to install any program off of GitHub via their REST API. Parm directly downloads binaries provided by GitHub repository releases and includes niceties such as symlinking binaries to PATH and checking for updates.

This means that Parm

- has zero root access & zero required dependencies
- requires no additional package maintainers
- receives new versions upstream instantly.
- is incredibly lightweight

### Demo

### Disclaimers
> [!CAUTION]
> Parm uses the GitHub REST API to find and install packages. Theoretically, this means you can install any program off of GitHub, so ***YOU*** are responsible for the packages you install, since I don't maintain a registry of vetted packages.

> [!WARNING]
> Unlike most package managers, Parm does NOT automatically resolve/install dependencies for you. This is a limitation of GitHub and the program's design. The current behavior is to use a tool such as `objdump` or `otool` to search for dynamically linked dependencies and tell you about them, but there is no way to automatically install them as of right now.

> [!NOTE]
> Parm is *not* intended to replace your system/OS-level package manager (think apt, pacman, or anything that can install low-level libraries, tools, or services). In general, it is designed to be a complement to your current package manager, as it is meant to install more high-level, user-facing applications.

## Quick Start
To install Parm on Linux/macOS: run the following command:
```sh
curl -fsSL https://raw.githubusercontent.com/yhoundz/parm/master/scripts/install.sh | sh
```

Windows is currently not fully supported, but will be coming soon. You can install the binaries manually in the releases tab.

To use parm:
```sh
parm install <owner>/<repo> # installs a package

parm remove <owner>/<repo> # uninstalls a package

parm update <owner>/<repo> # to update a package
```

For more detailed install instructions, see [Installation](#install).
For more detailed documentation, go to [Usage](#usage) or the [docs](#/docs/docs.md)

## Pre-requisites
1. *(optional)* Must have `objdump` on Linux or `otool` on macOS installed and added to PATH
    - Parm may use these tools to search for potential dependencies on installed programs. 
        - Parm will not try to search for dependencies on Windows at the moment.
        - On Linux, `ldd` is not used since some implementations of it may execute the program to find dependencies. This is more accurate than using `objdump` or `readelf`, but poses a bigger security risk given the scope and design of this project.
    - `objdump` and `otool` should already be installed and added to PATH on your machine. You can check this by running
    ```sh
    which objdump
    ```
    or
    ```sh
    which otool
    ```
2. *(optional)* Must have a **free** GitHub personal access token with access to PUBLIC repositories. Go [here](#adding-a-github-personal-access-token) to find out how to add an access token.

## Install
To install Parm on Linux/macOS: run the following command:
```sh
curl -fsSL https://raw.githubusercontent.com/yhoundz/parm/master/scripts/install.sh | sh
```

You can also set the `GITHUB_TOKEN` option to use your GitHub API Key to bypass rate limits:
```sh
GITHUB_TOKEN=YOUR_TOKEN curl -fsSL https://raw.githubusercontent.com/yhoundz/parm/master/scripts/install.sh | sh
```

In addition, you can also specify that the `GITHUB_TOKEN` you pass in will the same token you want to use for Parm by setting `WRITE_TOKEN=1`
```sh
GITHUB_TOKEN=YOUR_TOKEN WRITE_TOKEN=1 curl -fsSL https://raw.githubusercontent.com/yhoundz/parm/master/scripts/install.sh | sh
```

Windows is currently not fully supported, but will be coming soon. You can install the binaries manually in the releases tab.

## Adding a GitHub Personal Access Token
> [!IMPORTANT]
> If you DON'T have/want to use a GitHub personal access token, then you will be limited to 60 requests/hr instead of the 5000+ requests/hr with an API key. 
> There is nothing I can do about this and this is a limitation of the program's design and the GitHub API.

1. Add a personal access token (classic) by following [this guide](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic).
    - You can use a fine-grained personal access token if you want, but this is not tested properly. Check out the guide for that [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token).
2. Add the API key to your shell environment:
```sh
echo 'export GH_TOKEN=<your_token_here> >> ~/.bashrc'
```

- You can substitute `GH_TOKEN` with `GITHUB_TOKEN` or `PARM_GITHUB_TOKEN`
- You can also substitute `~/.bashrc` with your shell environment of choice (e.g. `~/.zshrc`)

3. Parm will automatically use your token from your shell's environment variable.
4. If Parm does not detect your token from your shell, you can set a fallback in the `github_api_token_fallback` in your config, which will be in `$XDG_CONFIG_HOME/parm/config.toml`

To use the `github_api_token_fallback` instead, run the following command (with parm installed):
```sh
parm config set github_api_token_fallback=<token>
```
5. If no token is found, it will default to non-authenticated GitHub REST API usage, which defaults to 60 requests/hr.

## Usage

| Command  | Flags | Description |
| ------------- | -------------- | -------------- |
| `install` | `--release, --pre-release, --asset, --strict, --no-verify` | Installs a package. |
| `uninstall` |  | Uninstalls a package. |
| `update` | `--strict` | Updates a package. |
| `list` |  | Lists the currently installed packages. |
| `config` |  | Prints out the current config in parm's `config.toml` file. |
| `config set` |  | Sets a `key=value` pair for a configuration setting. |
| `config reset` | `--all` | Resets a `key=value` config setting back to its default. |
| `info` | `--get-upstream` | Retrieves information about a certain package. |

For more detailed documentation, see the [docs](/docs/usage.md).

## Contributing
Parm is in a very early state, so any and all PRs are welcome. If you want to contribute to a new feature not already on the [roadmap](/docs/roadmap.md), please [create an issue](https://github.com/yhoundz/parm/issues/new) first, or check if an issue has already been created for it.

Before making a contribution, read over the [contributing guidelines](.github/CONTRIBUTING.md) as well as the [code of conduct](.github/CODE_OF_CONDUCT.md).

## Adding Packages to Parm
If you want to make your Github Repo to be installable via Parm, ensure the following:
- Your repository is PUBLIC.
- Your repository has at least ONE release with an associated release tag.
- Your repository release tag names follow semver (semantic versioning).
- Your repository is meant to be run on Windows, macOS, or Linux.
- Your repository release assets actually contain binaries
	- This is important, since some GitHub repositories don't actually do this (e.g. tmux/tmux only provides source code in their releases).

That's it! Your program is now compatible with Parm. To ensure maximized compatibility, I strongly recommend the following:
- Asset names must follow the convention: \<program-name>-\<OS-name>-\<arch-name>.\<file-extension>
	- For example, "parm-linux-amd64.tar.gz" follows this convention.
- Ensure that your program has no external dependencies, or requires no dependencies that aren't usually pre-installed on most users' machines.
	- This is because Parm's dependency resolution is very weak right now, due to how the program was designed, and how it's quite difficult to find dependencies programmatically without being given them explicitly (like how other package managers do this).
	- As a rule of thumb, if your program is statically linked, you should be good to go.

Parm tries its best to adhere to current conventions of open-source software when parsing repositories. If there's something I'm missing, please [create an issue](https://github.com/yhoundz/parm/issues/new).

## Acknowledgements
Parm was created using the [Go programming language](https://go.dev/) and the [cobra](https://cobra.dev/) CLI framework.

While not the direct inspirations for Parm, here are some projects that helped shape Parm's development:
- [homebrew](https://brew.sh/)
- [asdf](https://asdf-vm.com/)
- [lazy.nvim](https://lazy.folke.io/).

Similar projects:
- [ubi](https://github.com/houseabsolute/ubi)
- [eget](https://github.com/zyedidia/eget)
