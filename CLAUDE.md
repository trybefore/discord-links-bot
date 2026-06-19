# GitHub context

Host: https://github.com
Repo: trybefore/discord-links-bot

Read issues/PRs with the `gh` CLI (authenticated) or the public API via curl.
- Issue:   `gh issue view <n>`  /  /repos/trybefore/discord-links-bot/issues/{n}
- PR:      `gh pr view <n>`     /  /repos/trybefore/discord-links-bot/pulls/{n}
- Diff:    `gh pr diff <n>`     /  /trybefore/discord-links-bot/pull/{n}.diff

## Writing to the repo

`git push` works over HTTPS, and the `gh` CLI is authenticated for GitHub
operations (PRs, issues, comments). Common commands:
- `gh pr create --base <branch> --head <branch> --title <text> --body <text>` (`--fill` to autofill from commits)
- `gh pr view <n>`, `gh pr comment <n>`, `gh issue ...`

Still only commit/push/open PRs when I ask. When on the default branch, branch first.
