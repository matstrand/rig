# rig: a Gas Town inspired workflow without the overhead

Like many others, I've been enamored by the Gas Town metaphors for the software engineering workflow. Crew working on rigs, consuming guzzoline, slinging work to polecats... yeah, this is right up my alley.

The metaphors capture the way I want to think about how work gets done. Unfortunately, Gas Town itself isn't quite mature enough for me to adopt as my daily workflow.

## Why not Gas Town?

To be clear none of this is a dis or dunk on Gas Town - it's an ambitious project that is going to take time to mature. This is a snapshot of some of the problems that I observe today in Feb 2026.

**It's buggy.** It's a complicated system with bugs throughout. The Deacon and Witness run in loops consuming copious tokens to complete predefined tasks repeatedly. One bug that plagues it: these actors are supposed to have exponential backoff built in, but it's not always respected. You can easily spend $20/hr on the Deacon just trying to remember what it's supposed to do.

**It's expensive.** I don't mind the expense of a system that produces real value. But much of Gas Town's expense is overhead - autonomous agents running patrol loops whether or not there's work to do. Even if I can swallow the expense I'd still have a hard time with the social aspect of burning energy needlessly.

**It's opinionated in ways that don't fit my workflow.** Merging to main is almost a requirement, which is incompatible with PR-based workflows. The polecat workflow encodes `go test` as a requirement for success, and AFAIK that formula isn't yet parameterizable. You can fix it by forking and maintaining your own version, but things are moving fast over there and keeping up with main takes daily diligence.

**Beads** are difficult to manage and the polecats create them in unexpected ways. Duplicates are common and work is duplicated. Sometimes you don't even notice a swarm of polecats created 10 new issues to fix tests on main until they've all been processed and you've blown the budget. You end up with `.beads` directories everywhere. It's unclear what's in JSONL and what's in SQLite. The gastown maintainers are aware and they're working on migrating the backend to dolt. It's quite complicated.

**The mail sub-system** uses tmux send-keys to send messages between the claude codes. That's fine unless you're in the middle of typing out a big prompt to the mayor when he gets mail. When that happens whatever you have written gets sent to Claude along with the mail notification.

## So what is this?

I like these concepts too much to wait for Gas Town to mature. I need to move fast today, not in a year. And Gas Town is full of good ideas about how to move fast.

So I've integrated some of those concepts into this project. It gets its inspiration from Gas Town but basically nothing else. Whereas Gas Town is as complicated as necessary, this repo is as simple as possible. And it achieves many of the same goals with far less overhead.

## What got ported

* Rigs provide multiple workspaces for a repo using git worktrees.
* Crew have their own quarters, i.e. their own workspace.
* tmux is the primary interface.
* You can define work and sling it to an ephemeral crew member called a polecat.
* The current assignment is stored on a hook.
* You're not supposed to watch the polecats work.

## What is different

Well a lot is different but let me just call out a few things.

* Intermediate work is stored as commits on a feature branch.
* When work completes you review it and push the feature branch yourself.
* You write tickets in the form of markdown files in a `work/` directory organized by feature.
* Progress is tracked through a checklist in a markdown file in the same directory.

The workflow is basically:

1. Run `rig work create <feature-name>`. It creates a feature branch and scaffolds your spec.
2. Fill in the spec file: define your requirements. Use your favorite LLM to help. Commit it.
3. Sling the requirements to a polecat: `rig sling <feature-name>`. This creates a new workspace and tmux session that automatically begins working on the feature. As it progresses it updates `work/<feature-name>/progress.md`.
4. Check on status with `rig work status`: this shows you where work is at on all your rigs.
5. When work completes open the polecat and click the link to open a pull request. If you are happy with it, create the PR. If not, request changes until you are happy with it.
6. All done? Clean up the polecat: `rig crew remove <polecat_name>`. It will kill the session and prompt you to delete the local branch.
