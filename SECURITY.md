# Security

## Scope

MK-Link is a local Windows filesystem utility. Its most sensitive operation is creating a filesystem link at a user-selected destination.

## Security goals

1. Never execute user-controlled paths through a shell.
2. Never overwrite an existing destination implicitly.
3. Reject ambiguous or unsafe Windows link names.
4. Revalidate the operation in Go immediately before the native filesystem call.
5. Keep the web frontend limited to the application origin and bound Go methods.
6. Keep production browser context menus and developer tools disabled.

## Threat model

The application assumes the local user intentionally chooses the target and destination. It does not treat arbitrary filesystem paths as trusted merely because they came from the UI.

The UI is not a security boundary. Important validation happens in Go.

## Link-specific considerations

- Symbolic links can reference remote paths. This is a documented Windows capability and is only created after explicit user selection.
- Hard links are file-only and must be on the same volume. Windows enforces this at the native API boundary.
- Junctions are directory-only and MK-Link rejects UNC/network targets.
- Recursive directory-link destinations are rejected to avoid creating self-referential directory trees.

## Reporting

Do not disclose a security issue publicly before the maintainer has had an opportunity to investigate and release a fix. Include reproduction steps, affected version, Windows version, and the smallest possible proof of impact.
