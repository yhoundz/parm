---
icon: lucide/terminal
---

# Installing a Release

To install the latest stable release of a package, run

```sh
parm install <owner>/<repo>
```

For example:

```sh
parm install yhoundz/parm
```

> [!WARNING]
> Don't actually try to install Parm using Parm. It is unsupported and may or may not work how it was intended. Use the install script in the README instead.

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

Currently, Parm uses a naive text-matching scoring algorithm on asset names to determine which asset to download. However, this algorithm can be inaccurate and may not download the correct asset if the asset name is ambiguous.

To get around this, specify the asset name with the `--asset` flag when installing with the `--release` or `--pre-release` flag(s). Here is an example installing tmux, which has ambiguous asset names:

```sh
parm install tmux/tmux --release 3.5a --asset tmux-3.5a.tar.gz
```

An asset name is ambiguous if the algorithm cannot detect the intended architecture or OS the asset is intended for. For example, the algorithm will correctly detect the OS/arch for "parm-linux-x86_64.tar.gz" or "parm-macos-arm64.tar.gz", but will not detect the intended OS/arch name for "tmux-3.5a.tar.gz".

By default, Parm will also verify the downloaded tarball/zipball once it has been downloaded by generating a sha256 hash from the installed tarball and comparing it to the sha256 hash provided by the release asset upstream. To skip this verification, use the `--no-verify` flag:

```sh
parm install yhoundz/parm --no-verify
```

More options for checksum verification will be added in later versions.

## Installing a Pre-Release

You can install the latest pre-release as follows:

```sh
parm install yhoundz/parm --pre-release
```

By default, installing a pre-release will actually just install the latest possible release, which may be a stable release. For example, if a GitHub repository has a pre-release labelled "v2.0.0-beta", and the latest possible release is "v2.0.0", then it will install "v2.0.0" instead, even though it isn't a pre-release. "Pre-release" is just an umbrella term for installing the cutting-edge releases.

If you _strictly_ want to install pre-releases, use the `--strict` flag. Note that the `--strict` flag only works when using `--pre-release`:

```sh
parm install yhoundz/parm --pre-release --strict
```

---

# Updating a Package

To update a package, you can run the following command:

```sh
parm update <owner>/<repo>
```

Like the `install` command, if the package in question is on the pre-release channel, then you can use the `--strict` flag to only install pre-release versions, and not the latest cutting-edge version:

```sh
parm update yhoundz/parm --strict # assuming this is on the pre-release channel
```

---

# Uninstalling a Package

To remove/uninstall a package, you can run the following command:

```sh
parm remove <owner1>/<repo1> <owner2>/<repo2> ...
```

You can also use the `uninstall` command too if you wish; it is functionally the exact same as the `remove command`:

```sh
parm uninstall <owner>/<repo> ...
```

---

# Listing Installed Packages

You can list the currently installed packages with:

```sh
parm list
```

Due to the current implementation's lack of caching, this will likely be pretty slow, but fixes are planned for v0.2.0.

---

# Configuration

Parm's config file is in `$XDG_CONFIG_HOME/parm/config.toml`. If it can't find the `$XDG_CONFIG_HOME` environment variable, it will default to `$HOME/.config/parm/config.toml`

You can list out the current contents of the config file by running

```sh
parm config
```

This will print out the current configuration settings, though it will omit some such as GitHub personal access token environment variables. Do not that the `config` command will **NOT** omit the `github_api_token_fallback` configuration setting, so if you store your API key here, it will get printed out in its entirety. There will likely be a fix for this in future versions.

## Setting Configuration Options

You can set configuration options by running the `config set` subcommand, followed by a `key=value` pair for configuration settings you want to set. Ensure there are no spaces in between `key` and `value` (i.e. `key = value` is wrong).

```sh
parm config set key1=value1 key2=value2 ...
```

Parm comes with a set of default options. If for some reason you messed up your config, you can reset the config options using the `config reset` subcommand.

```sh
parm config reset key1 key2 ...
```

If you want to reset all configuration options back to their default, use the `--all` flag. Note this won't allow any additional arguments.

```sh
parm config reset --all
```

# Retrieving Package Information

To retrieve certain information about a package, use the `info` command.

```sh
parm info yhouhdz/parm
```

Here is a sample output of what it may look like:

```md
Owner: yhoundz
Repo: parm
Version: v0.1.0
LastUpdated: 2025-10-10 04:50:27
InstallPath: /home/user/.local/share/parm/pkg/yhoundz/parm
```

This displays most fields written to the manifest file upon installation. The full manifest file for a package, go to `$XDG_DATA_HOME/parm/pkg/<owner>/<repo>/.curdfile.json`

If you want more detailed information on a package, you can instead look at its upstream information by using the `--get-upstream` flag.

```sh
parm info yhoundz/parm --upstream
```

A sample output would look like this:

```md
Owner: yhoundz
Repo: parm
Version: v0.1.0
LastUpdated: 2025-10-03 22:34:28
Stars: 67
License: GPL-3.0 license
Description: Install any program from your terminal.
```

The information displayed will likely be tweaked and is not final at the moment.
