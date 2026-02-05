# Roadmap

> [!IMPORTANT]
> [Here](https://github.com/users/yhoundz/projects/2) is the GitHub Projects related to Parm. I will try to keep this file and the Projects board in sync with each other, but there are no promises regarding that. Assume if that item is *not* on the GitHub Projects, then it is not being worked on, or there are no plans to work on it.

Parm is still in a very early state, and breaking changes are to be expected. Additionally, a lot of features may not be implemented yet, or working as expected. If you would like to propose a new feature, [create a new issue](https://github.com/yhoundz/parm/issues/new).

Below are a list of planned features and improvements:

## Planned Features/Improvements

### Planned for v0.2.0
See the milestone [here](https://github.com/yhoundz/parm/issues/43)

### General Improvements
- Vetting/replacing AI-generated tests with better ones, more test coverage.
- Logging to a file, both informational and error logging.
	- Replacing fmt.Println(), logging instead which will write to stdout and a file
- Implement GraphQL (githubv4) support
- Caching API calls or expensive operations (like listing installed packages)

## Planned for Later Versions
- Better version management: Entails being able to install multiple versions at once and switching between them easily.
- Search feature, allowing users to search for repositories through parm via the GraphQL API.
- Shell autocompletion (for --asset flag, uninstalling packages, updating packages)
- "doctor" command to verify everything works as intended.

## To be Determined
- Add migrate command if user changes bin or install dir in config.
- Resolve potential collisions between installed repos and symlinked binaries if two "owners" have packages with the same name.
- Parsing different kinds of binary files (not just ELF/Macho/PE)
- Parse binaries for dependenices myself without using `objdump` or `otool -L`.
	- The current solution is to parse the output of `objdump` and `otool -L`, but their outputs are designed to be human-readable and not machine-readable. Implementing this would mitigate that.
	- Better dependency resolution; implement an algorithm mimicking the linker's shared library searching algorithm to only find dependencies that are NOT currently on the user's system.
