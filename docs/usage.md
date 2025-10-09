# Installing a Release
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

## Installing a Pre-Release

You can install the latest pre-release as follows:
```sh
parm install yhoundz/parm --pre-release
```

--- 

# Updating a Package

To update a package, you can run the following command:
```sh
parm update <owner>/<repo>
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
