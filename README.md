# Parm 🧀

Parm is a package/repository manager that allows you to install any program binary from github and will manage it for you.

> [!CAUTION]
> Parm is currently in a pre-release state. Expect breaking changes and bugs.

## Install:
TODO

## Pre-requisites:
1. Must have git installed and added to PATH
2. Must have a **free** GitHub personal access token with access to PUBLIC repositories (optional)

## Add GitHub API Key/Personal Access Token:
> [!IMPORTANT]
> If you DON'T have/want to use a GitHub personal access token, then you will be limited to 60 requests/hr instead of the 5000+ requests/hr with an API key. 
> There is nothing I can do about this and this is a limitation of the program's design and the GitHub API.

Parm uses the GitHub REST API to find and install packages. Theoretically, this means you can install any program off of GitHub, so ***YOU*** are responsible for the packages you install, since I don't maintain a registry of vetted packages.

### Add Personal Access Token:
1. Add a personal access token (classic) by following [this guide](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic).
    - You can use a fine-grained personal access token if you want, but this is not tested properly. Check out the guide for that [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token).
2. Add the API key to your shell environment:

```bash
echo 'export GH_TOKEN=<your_token_here> >> ~/.bashrc'
```

- You can substitute `GH_TOKEN` with `GITHUB_TOKEN` or `PARM_GITHUB_TOKEN`
- You can also substitute `~/.bashrc` with your shell environment of choice (e.g. `~/.zshrc`)

3. Parm will automatically use your token from your shell's environment variable.
4. If Parm does not detect your token from your shell, you can set a fallback in the `github_api_token_fallback` in your config, which will be in `$XDG_CONFIG_HOME/parm/config.toml` 
5. If no token is found, it will default to non-authenticated GitHub REST API usage, which defaults to 60 requests/hr.

## Usage:

### Installing Releases (Recommended):
To install the latest stable release of a package, run
```bash
parm install <owner>/<repo>
```

For example:
```bash
parm install yhoundz/parm
```

If you want, you can also specify the full https/ssh link for installation:
```bash
parm install https://github.com/yhoundz/parm.git
```
or
```bash
parm install git@github.com:yhoundz/parm.git
```


If you want to install a specific version of a package, you can specify with the --release flag
```bash
parm install yhoundz/parm --release v0.1.0
```

This will install the specific GitHub release assoicated with the release tag.
You can also use the "@" keyword as a shorthand, as follows:
```bash
parm install yhoundz/parm@v0.1.0
```

You can also install directly from source, but you will have to resolve dependencies and build it yourself. Not recommended unless you know what you're doing.
This is effectively the same as running "git clone" (though it actually installs the source code as it was at the corresponding release).
```bash
parm install yhoundz/parm --release v0.1.0 --source
```

#### Installing a Pre-Release:

You can install the latest pre-release as follows:
```bash
parm install yhoundz/parm --pre-release
```

And of course, you can specify if you want to install from source instead:
```bash
parm install yhoundz/parm --pre-release --source
```
