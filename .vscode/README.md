# VSCode Configuration

If you opt to use VSCode, this package contains some optional workspace
settingsthat will improve your development experience in this repo.

The following extensions are recommended:

- ESLint - https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint
- JS + TS Nightly - https://marketplace.visualstudio.com/items?itemName=ms-vscode.vscode-typescript-next
- Vetur - https://marketplace.visualstudio.com/items?itemName=octref.vetur 

I recommend using Auto-format on save according to our ESLint configuration, 
disabling VSCode's default formatting for Typescript (in particular).

Vetur provides Vue Auto-Completion + 'view definition' hints, which
are exceptionally helpful. However, Vetur's validation doesn't take
into account our eslint overrides. This means globals get complained
about, which is obnoxious. Thus, I recommend turning of vuteur's validation, 
while still enabling their extension for cross checking templates, code 
completion, and definition links, all of which I've found helpful.