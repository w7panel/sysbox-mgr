# Changelog

## Unreleased

- Add explicit `nested-identity` registration and allocation. A manager in a
  non-initial user namespace can return `0:0:65536` without reading or
  modifying L1 `/etc/subuid` or `/etc/subgid`, and multiple L2 containers may
  reuse that identity mapping. Network namespace updates replace the original
  CNI inode after runc migrates the sandbox into its child-owned netns.
