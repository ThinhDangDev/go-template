# Repository Guidelines

## Release Discipline

- This repository is release-driven. A change is not complete until the commit is pushed and a matching semantic version tag is pushed when the change is meant to ship.
- Use annotated tags only: `git tag -a vX.Y.Z -m "vX.Y.Z"`.
- This matters because users install the CLI with `go install github.com/ThinhDangDev/go-template/cmd/go-template@latest`, so published tags control what users actually receive.

## Version Bump Rules

- Bump `patch` for docs-only updates, tests, refactors, CI/build tweaks, dependency bumps, and bug fixes that do not change the CLI contract or generated project surface.
- Bump `minor` for any new CLI capability, new scaffolded API, new template module, new generated project behavior, or any template change that adds visible functionality.
- While the project is still below `v1.0.0`, treat breaking CLI/template contract changes as at least a `minor` bump.
- After `v1.0.0`, use `major` for breaking changes.
- If the right bump is ambiguous, call it out and default to the safer higher bump.

## Push Rules

- Push the commit first, then push the release tag that points to the same commit.
- If working on `main`, the release tag must point at `HEAD`.
- If working on a non-`main` branch, push the branch first and only create the release tag once the tagged commit is on `main`, unless the user explicitly asks for a prerelease tag.
- After pushing, verify the remote branch and the remote tag resolve to the expected commit.

## Auth Fallback

- If SSH push fails because the loaded key does not have permission, use GitHub CLI HTTPS credentials for push operations:
  `git -c credential.helper='!gh auth git-credential' push https://github.com/ThinhDangDev/go-template.git HEAD:main`
- Use the same credential helper pattern for pushing tags and for remote verification when needed.

## Agent Expectation

- For every completed change in this repo, agents should explicitly state the chosen version bump, create the tag, push the commit, push the tag, and confirm the remote ref that was published.
