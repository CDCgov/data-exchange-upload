# DEX TUSD Go Server 
A resumable file upload server for OCIO Data Exchange (DEX)

## Folder structure
Repo is structured (as feasible) based on the [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

## References
- Based on the [tus](https://tus.io/) open protocol for resumable file uploads
- Based on the [tusd](https://github.com/tus/tusd) official reference implementation

## VS Code 
.vscode/launch.json
```js
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            // "program": "${fileDirname}"
            "program": "cmd/main.go",
            "cwd": "${workspaceFolder}",
            "args": [
                "-env", "local"
            ]
        }
    ]
}
```

