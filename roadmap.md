# Roadmap

Parm is still in a very early state, and breaking changes are to be expected. Additionally, a lot of features may not be implemented yet, or working as expected. If you would like to propose a new feature, [create a new issue](https://github.com/yhoundz/parm/issues/new).

Below are a list of planned features and improvements:

## Planned for v0.2.0
- Logging to a file, both informational and error logging.
- Allow users to be able to choose which asset to release if a direct match isn't found.
- Caching API calls or expensive operations (like listing installed packages)
- Better version management: Entails being able to install multiple versions at once and switching between them easily.
- Boostrapping: Allowing the user to update Parm itself without having to rerun the install script again and do it from within the CLI.
- Switching release channels: Allow the user to switch between the "Release" and "Pre-release" channels, changing how the update command behaves.

## Planned for Later Versions
- Shell autocompletion.
- Checksum comparison (i.e. the user can input a checksum as a flag, and Parm will compare the user's checksum to the checksum of the downloaded tarball)

## To be Determined
- None
