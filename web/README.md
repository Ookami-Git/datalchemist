# datalchemist-vue

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur) + [TypeScript Vue Plugin (Volar)](https://marketplace.visualstudio.com/items?itemName=Vue.vscode-typescript-vue-plugin).

## Customize configuration

See [Vite Configuration Reference](https://vitejs.dev/config/).

## Project Setup

This package is managed with **pnpm**. The version is pinned by the
`packageManager` field in `package.json`, so the easiest way to get the right
one is through corepack (bundled with Node):

```sh
corepack enable pnpm
```

```sh
pnpm install
```

Use `pnpm install --frozen-lockfile` in CI and for reproducible installs: it
fails instead of silently updating `pnpm-lock.yaml`.

### Compile and Hot-Reload for Development

```sh
pnpm dev
```

### Compile and Minify for Production

```sh
pnpm build
```

### Run the tests

```sh
pnpm test
```

## Dependency policy

`pnpm-workspace.yaml` enforces a few supply-chain rules; they are there on
purpose, so read the comments before relaxing them:

- `minimumReleaseAge: 1440` — a version published less than 24h ago is never
  installed. Compromised releases are usually pulled within hours, so this is
  the main protection against a poisoned publish. If you *must* take a fresh
  release, add a narrow entry to `minimumReleaseAgeExclude` rather than
  lowering the global value.
- `allowBuilds` — only the listed packages may run install/build scripts.
  A new dependency that needs one will fail the install until it is reviewed
  and added, which is what stops `preinstall`-style payloads.
- `overrides` — pins patched versions of transitive dependencies that cannot be
  fixed by bumping a direct dependency. Re-check these when bumping deps.

Note that pnpm uses an isolated `node_modules`, so a package must be declared in
`package.json` to be importable. Relying on a transitive dependency (a "phantom
dependency") fails the build rather than working by accident.
