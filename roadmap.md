# Roadmap

Parm is still in a very early state, and breaking changes are to be expected. Additionally, a lot of features may not be implemented yet, or working as expected. If you would like to propose a new feature, [create a new issue](https://github.com/yhoundz/parm/issues/new).

Below are a list of planned features and improvements:

## Planned for Completion by v0.2.0
- Logging to a file, both informational and error logging.
- Allow users to be able to choose which asset to release if a direct match isn't found.
- Better version management: Entails being able to install multiple versions at once and switching between them easily.
- Boostrapping: Allowing the user to update Parm itself without having to rerun the install script again and do it from within the CLI.
- Switching release channels: Allow the user to switch between the "Release" and "Pre-release" channels, changing how the update command behaves.
- Resolve potential collisions between installed repos and symlinked binaries if two "owners" have packages with the same name.

## Planned for Later Versions
- Shell autocompletion (for --asset flag, uninstalling packages, updating packages)
- Add migrate command if user changes bin or install dir in config.

## To be Determined
- Caching API calls or expensive operations (like listing installed packages)
