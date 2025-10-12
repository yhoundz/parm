# Roadmap

Parm is still in a very early state, and breaking changes are to be expected. Additionally, a lot of features may not be implemented yet, or working as expected. If you would like to propose a new feature, [create a new issue](https://github.com/yhoundz/parm/issues/new).

Below are a list of planned features and improvements:

## Planned for Completion by v0.2.0

### Feature Improvements
- Boostrapping: Allowing the user to update Parm itself without having to rerun the install script again and do it from within the CLI.
- Switching release channels: Allow the user to switch between the "Release" and "Pre-release" channels, changing how the update command behaves.
- Add verification levels with the --verify flag
	- Flags would then be --verify, --no-verify (in v0.1.0) and --sha256
	- level 0: No verification
	- level 1 (default):
		* Generated hash gets compared to upstream post-download (pre-extraction)
	- level 2:
		* User-provided hash gets compared to upstream pre-download
		* Generated hash gets compared to upstream post-download (pre-extraction)
	- level 3:
		* User-provided hash gets compared to upstream pre-download
		* Generated hash gets compared to upstream post-download (pre-extraction)
		* Computed hash of untarred/unzipped files gets compared to upstream post-extraction (this is moreso a file integrity check more than a security check)

### General Improvements
- Vetting/replacing AI-generated tests with better ones, more test coverage.
- Refactor CLI commands to be generated via a method, not statically
- Logging to a file, both informational and error logging.
	- Replacing fmt.Println(), logging instead which will write to stdout and a file
- Implement GraphQL (githubv4) support
- Caching API calls or expensive operations (like listing installed packages)

## Planned for Later Versions
- Better version management: Entails being able to install multiple versions at once and switching between them easily.
- Search feature, allowing users to search for repositories through parm via the GraphQL API.
- Shell autocompletion (for --asset flag, uninstalling packages, updating packages)
- Parse binaries for dependenices myself without using `objdump` or `otool -L`.
	- The current solution is to parse the output of `objdump` and `otool -L`, but their outputs are designed to be human-readable and not machine-readable. Implementing this would mitigate that.
	- Better dependency resolution; implement an algorithm mimicking the linker's shared library searching algorithm to only find dependencies that are NOT currently on the user's system.
- Allow users to be able to choose which asset to release if a direct match isn't found.

## To be Determined
- Add migrate command if user changes bin or install dir in config.
- Resolve potential collisions between installed repos and symlinked binaries if two "owners" have packages with the same name.
- Parsing different kinds of binary files (not just ELF/Macho/PE)
