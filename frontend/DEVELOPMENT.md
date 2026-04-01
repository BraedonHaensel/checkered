# Frontend Development

This document describes the recommended setup for frontend development in [Visual Studio Code](https://code.visualstudio.com/).

## Recommended Extensions

ESLint

- Analyze all files with `npm run lint`

GitLens

Path Intellisense

Prettier - Code formatter

- Format all files with `npx prettier . --write`

Tailwind CSS IntelliSense

## Recommended VSCode Settings

Create a `.vscode/settings.json` file with the following configuration:

```json
{
    "editor.tabSize": 4,
    "editor.formatOnPaste": true,
    "editor.formatOnSave": true,
    "editor.formatOnSaveMode": "file",
    "files.insertFinalNewline": true,
    "editor.defaultFormatter": "esbenp.prettier-vscode"
}
```

Restart VSCode to apply the new configuration.
