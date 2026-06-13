# gotd documentation

Source for [gotd.dev](https://gotd.dev), the documentation website for [gotd](https://github.com/gotd/td) — a Telegram client in Go.

Built with [Docusaurus 3](https://docusaurus.io/), a modern static website generator.

## Installation

```
$ yarn
```

## Local Development

```
$ yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```
$ yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Generating reference docs

```
$ yarn gen:reference
```

Regenerates the API reference pages via `tools/docgen/generate.sh`.

## Deployment

```
$ GIT_USER=<Your GitHub username> USE_SSH=true yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.
