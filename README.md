# Parm 🧀

Parm is a **pa**ckage/**r**epository **m**anager that allows you to install any program binary from a github release and will manage it for you.

> [!CAUTION]
> Parm is currently in a pre-release state. Expect breaking changes and bugs.

**Table of Contents:**
1. [Why Parm?](#why-parm)
2. [Pre-requisites](#pre-requisites)
3. [Install](#install)
4. [GitHub Personal Access Token](#github-personal-access-token)
    - [Add Personal Access Token](#add-personal-access-token)
5. [Usage](#usage)
    - [Installing Releases](#installing-releases)
        - [Installing a Pre-Release](#installing-a-pre-release)
    - [Updating a Package](#updating-a-package)
    - [Deleting a Package](#deleting-a-package)
6. [Contributing](#contributing)

## Why Parm?
In short: if you like your current package manager, this isn't for you.
I originally wanted to build this because I use an older version of Linux, which often has a lot of outdated packages. As such, if I wanted to install a recent release, I would have to either build it from source, use homebrew, or find some alternative installation method. These are all viable solutions, but they were:

- Too cumbersome (in the case of building from source)
- Hard to work with (like homebrew, which makes it difficult to install any other version except the latest release version. Even then, the version you want may not be guaranteed to exist)
- Non-standardized (in the case of alternative install methods, such as esoteric install scripts)

I just wanted a single, unified way to manage my installed programs without having to deal with stale versions on my OS package manager or having to installing packages in a non-standardized way.

## Pre-requisites
1. Must have git installed and added to PATH
2. *(optional)* Must have a **free** GitHub personal access token with access to PUBLIC repositories

## Install
TODO

## GitHub Personal Access Token
> [!IMPORTANT]
> If you DON'T have/want to use a GitHub personal access token, then you will be limited to 60 requests/hr instead of the 5000+ requests/hr with an API key. 
> There is nothing I can do about this and this is a limitation of the program's design and the GitHub API.

Parm uses the GitHub REST API to find and install packages. Theoretically, this means you can install any program off of GitHub, so ***YOU*** are responsible for the packages you install, since I don't maintain a registry of vetted packages.

### Add Personal Access Token
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

## Contributing
Parm is in a very early state, so PRs are welcome. However, PRs should be related to bugs or unintended behavior, and no additional feature requests will be considered if they are not already on the [roadmap](#ROADMAP.md). If you would like to start discourse on a new feature, create an issue on GitHub.

Before making a contribution, read over the [contributing guidelines](#/contributing.md).
