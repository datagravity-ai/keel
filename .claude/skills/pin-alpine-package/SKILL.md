---
allowed-tools: Read, Edit, Grep, Bash, WebFetch, AskUserQuestion
argument-hint: '[alpine:<pkg> | <vulnerable-version> | <fixed-version>] [...more lines]'
description: Pin an Alpine package to a CVE-fixed version in the Dockerfile, with the correct package name and renovate annotation.
---

# Pin Alpine Package (CVE Security Patch)

This skill pins one or more Alpine packages to a security-fixed version in the Dockerfile
so containers resolve to the patched version. It is typically driven by a vulnerability
scanner (e.g. Aikido) that reports findings in the form:

```
alpine:<source-package> | <vulnerable-version> | <fixed-version>
```

For example: `alpine:openssl | 3.5.6-r0 | 3.5.7-r0`

The middle column is the **vulnerable** version, the last column is the **fixed** version. We
pin to the **fixed** version (the rightmost).

## Background — how package pinning works here

System packages are installed in the `Dockerfile` via `apk --no-cache add`. Pins use the
syntax `package=version`. CVE security pins are added to the `RUN apk --no-cache add` line,
each preceded by a renovate annotation comment so Renovate tracks future upgrades.

**Critical:** Alpine package names are usually straightforward (e.g. `openssl`, `libcrypto3`,
`libssl3`), but verify the actual package name that ships the fix. Use the Alpine package
database to confirm.

## Arguments

Parse `$ARGUMENTS` for one or more scanner lines of the form
`alpine:<pkg> | <vulnerable> | <fixed>`. There may be several, one per line. Free-text is also
fine (e.g. "pin openssl to 3.5.7-r0").

If `$ARGUMENTS` is empty, use `AskUserQuestion` to ask the user to paste the scanner
finding(s).

## Step 1: Parse each finding

For each line, extract:
- **package name** — the part after `alpine:` (e.g. `openssl`)
- **fixed version** — the **rightmost** version (e.g. `3.5.7-r0`)

Ignore the vulnerable (middle) version except to confirm we're upgrading, not downgrading.

## Step 2: Check whether it's already pinned

Read `Dockerfile`.

- If the package is **already pinned to the fixed version (or newer)**, report
  "already done" and skip it.
- If pinned to an **older** version, you'll update it in Step 4.

## Step 3: Resolve the correct package name

Determine the actual Alpine package name that ships the fix.

1. **Verify against the Alpine package index** using `WebFetch`:
   - Package search: `https://pkgs.alpinelinux.org/packages?name=<package>&branch=v3.22`
   - Package details: `https://pkgs.alpinelinux.org/package/v3.22/main/x86_64/<package>`

2. **Common mappings:**
   - `openssl` source produces: `openssl`, `libcrypto3`, `libssl3`
   - Usually pinning just `openssl` is sufficient as it pulls matching `libcrypto3`/`libssl3`

3. **Which packages to pin?** Pin what the scanner flagged. If the scanner flags the
   source package name, pin the main package (e.g. `openssl`). If it flags a specific
   library (e.g. `libcrypto3`), pin that specifically.

4. **Confirm the fixed version string is valid** before writing it. The pin must be an
   exact, real version or the build fails.

## Step 4: Apply the pin

Edit `Dockerfile`. Add a renovate annotation comment immediately before the `RUN apk` line
that installs the package, and add the package with version pin to that line:

```dockerfile
# renovate: datasource=repology depName=alpine_3_22/<package> versioning=loose
RUN apk --no-cache add ca-certificates <package>=<fixed-version>
```

Concrete example for `alpine:openssl | 3.5.6-r0 | 3.5.7-r0`:

```dockerfile
# renovate: datasource=repology depName=alpine_3_22/openssl versioning=loose
RUN apk --no-cache add ca-certificates openssl=3.5.7-r0
```

Rules:
- Format is exactly `package=version` (single `=`).
- The renovate `depName` uses `alpine_3_22/<package>` and `versioning=loose`.
- The datasource is `repology`.
- If updating an existing pin, change the version in place and keep its renovate comment.
- If multiple packages need pinning, add one renovate comment per package on separate lines
  before the `RUN` command, or split into multiple `RUN` lines.

## Step 5: Report and remind

Summarize what changed (which packages, old → new versions, which were already done).

**Always include this reminder:**

> The pin takes effect on the next Docker image build. The nightly cron job
> (`docker.yaml`) rebuilds the image daily, so the fix will be picked up within 24 hours.
> To deploy immediately, trigger the Docker workflow manually or push to master.

Then offer to open a PR (do not open one automatically).

## Common pitfalls

- **Wrong Alpine version in repology depName** — ensure it matches the Alpine base image
  version (currently `alpine_3_22` for `alpine:latest` which resolves to 3.22.x).
- **Guessing the version** — always verify the exact version string exists in the Alpine
  package repository; a wrong string fails the build.
- **Missing renovate annotation** — every CVE pin needs one so it's tracked for future
  upgrades.
- **Forgetting `-r<N>` suffix** — Alpine versions include a release suffix like `-r0`,
  `-r1` etc. Always include it.
