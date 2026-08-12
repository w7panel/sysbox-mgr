# Changelog

## Unreleased

- Let the experimental nested runtime allocate and retain a `1:65535` subuid
  and subgid range. The normal 65,536-ID block includes an ID unavailable in
  the outer user namespace and caused inner runtime `chown` operations to
  fail with `EINVAL`.
