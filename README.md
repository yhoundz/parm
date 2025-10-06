# Parm 🧀

Parm is a **pa**ckage/**r**epository **m**anager that allows you to install any program binary from a github release and will manage it for you. It's meant to be lightweight, has zero root access, and requires no dependencies.

> [!WARNING]
> Parm is currently in a pre-release state. Expect breaking changes and bugs.

**Table of Contents:**
1. [Introduction](#introduction)
    - [What is Parm?](#what-is-parm)
    - [Motivation](#motivation)
    - [Disclaimers](#disclaimers)
2. [Pre-requisites](#pre-requisites)
3. [Installation](#install)
4. [GitHub Personal Access Token](#adding-a-github-personal-access-token)
5. [Usage/Documentation](#usage)
    - [Installing Releases](#installing-releases)
        - [Installing a Pre-Release](#installing-a-pre-release)
    - [Updating a Package](#updating-a-package)
    - [Deleting a Package](#deleting-a-package)
6. [Contributing](#contributing)
7. [Acknowledgements](#Acknowledgements)

## Introduction
**TL;DR**: If you like your current package manager, this isn't for you. If not, keep reading or [install](#install) Parm.

### What is Parm?
Parm is a cross-platform package manager that allows you to install any program off of GitHub via their REST API. Parm directly downloads binaries provided by GitHub repository releases and includes niceties such as symlinking binaries to PATH and checking for updates.

This means that Parm:

- has zero root access & zero required dependencies
- requires no additional package maintainers
- receives new versions upstream instantly.
- is incredibly lightweight

### Motivation
My builtin package manager (apt) often has a lot of outdated packages. As such, if I wanted to install a recent release, I would have to either build it from source, use homebrew, or find some alternative installation method. While they're all acceptable solutions, I felt they were:

- Too cumbersome (in the case of building from source)
- Hard to work with (like homebrew, which can be difficult to install any other version except the latest release version. Even then, the version you want may not even be available)
- Non-standardized (in the case of alternative install methods, such as esoteric install scripts)

I just wanted a single, unified way to manage my installed programs without having to deal with stale versions on my OS package manager or having to installing packages in a non-standardized way.

### Disclaimers
> [!CAUTION]
> Parm uses the GitHub REST API to find and install packages. Theoretically, this means you can install any program off of GitHub, so ***YOU*** are responsible for the packages you install, since I don't maintain a registry of vetted packages.

> [!NOTE]
> Parm is *not* intended to replace your system/OS-level package manager (think apt, pacman, or anything that can install low-level libraries, tools, or services). In general, it is closer to programs like homebrew, as it is meant to install more high-level, user-facing applications such as neovim.


## Pre-requisites
1. *(optional)* Must have `ldd` on Linux or `otool` on macOS installed and added to PATH
    - Parm may use these tools to search for potential dependencies on installed programs. 
        - If these tools are not found or if there is an error invoking them, then Parm will fallback to a naive dependency search algorithm instead which may be inaccurate.
        - It is important to note that Windows machines do not have a similar tool out of the box and have a different linking process, so Parm won't try to resolve dependencies at all on Windows machine at the moment.
    - Both `ldd` and `otool` should already be installed and added to PATH on your machine. You can check this by running
    ```sh
    which ldd
    ```
    or
    ```sh
    which otool
    ```
2. *(optional)* Must have a **free** GitHub personal access token with access to PUBLIC repositories. Go [here](#adding-a-github-personal-access-token) to find out how to add an access token.

## Install
TODO

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
5. If no token is found, it will default to non-authenticated GitHub REST API usage, which defaults to 60 requests/hr.

## Usage

### Installing Releases
To install the latest stable release of a package, run
```sh
parm install <owner>/<repo>
```

For example:
```sh
parm install yhoundz/parm
```

If you want, you can also specify the full https/ssh link for installation:
```sh
parm install https://github.com/yhoundz/parm.git
```
or
```sh
parm install git@github.com:yhoundz/parm.git
```

Installing with no flags will automatically install the latest stable release of a package.
If you want to install a specific version of a package, you can specify with the --release flag:
```sh
parm install yhoundz/parm --release v0.1.0
```

This will install the specific GitHub release assoicated with the release tag.
You can also use the "@" keyword as a shorthand, as follows:
```sh
parm install yhoundz/parm@v0.1.0
```

#### Installing a Pre-Release

You can install the latest pre-release as follows:
```sh
parm install yhoundz/parm --pre-release
```

### Updating a Package

To update a package, you can run the following command:
```sh
parm update <owner>/<repo>
```

### Deleting a Package

To remove/uninstall a package, you can run the following command:
```sh
parm remove <owner1>/<repo1> <owner2>/<repo2> ...
```

You can also use the `uninstall` command too if you wish; it is functionally the exact same as the `remove command`:
```sh
parm uninstall <owner>/<repo> ...
```

## Contributing
Parm is in a very early state, so any and all PRs are welcome. If you want to contribute to a new feature not already on the [roadmap](ROADMAP.md), please [create an issue](https://github.com/yhoundz/parm/issues/new) first, or check if an issue has already been created for it.

Before making a contribution, read over the [contributing guidelines](CONTRIBUTING.md) as well as the [code of conduct](CODE_OF_CONDUCT.md).

## Acknowledgements
Parm was created using the [Go programming language](https://go.dev/) and the [cobra](https://cobra.dev/) CLI framework.

While not the direct inspirations for Parm, here are some projects that helped shape Parm's development:
- [homebrew](https://brew.sh/)
- [asdf](https://asdf-vm.com/)
- [lazy.nvim](https://lazy.folke.io/).
