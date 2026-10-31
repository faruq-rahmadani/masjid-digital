# components

The design system. Import from `@/components`; the spec is
[docs/design/design-system.md](../../../docs/design/design-system.md) and `/kit` (dev) shows each
component live.

| Folder          | Holds                                                                    |
| --------------- | ------------------------------------------------------------------------ |
| `actions/`      | Things you press: buttons, icon buttons, copy                            |
| `forms/`        | Inputs and choices, all wired through `Field`                            |
| `data-display/` | Read-only content: cards, lists, tables, badges, money                   |
| `feedback/`     | System status: callouts, toasts, empty/loading states, progress          |
| `overlays/`     | Layers above the page: sheets, menus, confirmations                      |
| `navigation/`   | Moving between screens and views: headers, tabs, links (uses the router) |
| `layout/`       | Page structure: hero, body, action bar, spacing                          |

Rules:

- No data fetching and no feature imports here; features compose these.
- Style only with tokens from `src/styles/index.css`. Accept `className` and merge it with `cn()`.
- New component? Add it to the right folder, the barrel (`index.ts`), `/kit` and the design-system doc.
