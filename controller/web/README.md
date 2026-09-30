# CYPHONIC cloud software controller: WEB

[![TypeScript](https://img.shields.io/badge/TypeScript-5.3.3-blue.svg)](https://github.com/microsoft/TypeScript/releases/tag/v5.3.5)
[![Next.js](https://img.shields.io/badge/Next.js-14.0.4-blue.svg)](https://github.com/vercel/next.js/releases/tag/v14.0.4)

[![web-reviewdog](https://github.com/Pluslab/cyphonic/actions/workflows/web-review.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/web-review.yaml)

## Next.js Project

This guide will walk you through cloning and starting a Next.js project from a GitHub repository. Before starting, make sure you have the following installed:

- Node.js (version ^20 or later)
- Git

## Getting Started

**WARNING: API server must be started in advance.**

### Open Project

First, entering the container containing the next.js project:

```sh
$ make web/sh
```

### Installing Dependencies

Executing the following command to install the project dependencies:

```sh
# make install
```

This will install all the required dependencies for the project, including Next.js.

### Starting the Development Server

Starting the development server by executes the following command:

```sh
# make run
```

This will start the development server on <http://localhost:3001>. You can now open this URL in your web browser to view the project.
