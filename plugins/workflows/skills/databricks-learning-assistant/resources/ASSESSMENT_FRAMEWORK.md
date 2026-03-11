# Assessment Framework

Assessment-through-action methodology for the Databricks Learning Assistant. Users prove understanding by deploying resources, running commands, and explaining their reasoning — not by answering quiz questions.

---

## Core Philosophy

- **Assessment = action, not quizzes.** The user demonstrates competence by doing real work.
- **Verify with CLI — always.** Every claim of success is verified by running a command. **Never skip verification.** If the user says "I did it" or "done," run the check. If the resource doesn't exist or the result doesn't match, the step is not complete — guide them to actually do it.
- **The assistant verifies, it does not do.** The only actions the assistant should take during assessment are verification commands (queries, CLI checks, file reads). The assistant must NOT create resources, run feature code, or complete steps on behalf of the user. The user does the work; the assistant confirms it was done correctly.
- **"Why" matters as much as "what."** Understanding the reasoning behind a choice is as important as getting the right result.
- **Low pressure.** This is learning, not an exam. Frame challenges as practice, not tests.

---

## Deployment Verification Patterns

Use CLI commands to verify the user's work. These are template patterns — substitute the actual resource names and profiles discovered during the session.

### Resource Existence Checks

```bash
# Verify a table exists and has expected properties
databricks tables get <catalog>.<schema>.<table> --profile <profile>

# Verify a cluster is running with expected config
databricks clusters get <cluster-id> --profile <profile>

# Verify a job was created
databricks jobs list --name <job-name> --profile <profile>

# Verify a warehouse is running
databricks warehouses get <warehouse-id> --profile <profile>

# Verify an app is deployed
databricks apps get <app-name> --profile <profile>

# Verify a pipeline exists
databricks pipelines get <pipeline-id> --profile <profile>

# Verify permissions were granted
databricks grants get <securable-type> <securable-name> --profile <profile>
```

### Data Verification Checks

```bash
# Verify data exists in a table
databricks sql query execute --query "SELECT COUNT(*) FROM <catalog>.<schema>.<table>" --warehouse-id <id> --profile <profile>

# Verify table properties
databricks sql query execute --query "DESCRIBE DETAIL <catalog>.<schema>.<table>" --warehouse-id <id> --profile <profile>

# Verify a specific configuration
databricks sql query execute --query "SHOW TBLPROPERTIES <catalog>.<schema>.<table>" --warehouse-id <id> --profile <profile>
```

### State Verification Checks

```bash
# Verify pipeline is running and healthy
databricks pipelines get <pipeline-id> --profile <profile> | grep state

# Verify job run completed
databricks runs get <run-id> --profile <profile> | grep state

# Verify serving endpoint is ready
databricks serving-endpoints get <endpoint-name> --profile <profile> | grep state
```

---

## Conceptual Question Templates

Use these "why" question patterns to assess understanding. Adapt to the specific feature and context — don't ask these verbatim.

### Architecture Decisions
- "Why did we use [choice A] instead of [choice B]?" (e.g., shared access mode vs single-user)
- "What would change if we needed to support [different requirement]?"
- "When would you choose [alternative approach] instead?"

### Configuration Reasoning
- "Why did we set [config option] to [value]?"
- "What would happen if we changed [config] to [different value]?"
- "Which of these settings is the most important for [use case] and why?"

### Failure Mode Understanding
- "If [component] goes down, what happens to [dependent system]?"
- "What's the first thing you'd check if [symptom] occurred?"
- "How would you tell the difference between [failure type A] and [failure type B]?"

### Trade-off Analysis
- "What are we trading off by choosing [approach]?"
- "If cost was the primary concern, what would you change?"
- "If performance was the primary concern, what would you change?"

Pick 1-2 questions relevant to what the user just learned. Don't ask all of them.

---

## Variation Challenge Patterns

After the guided walkthrough, challenge the user to apply what they learned independently. Adapt these patterns to the specific feature.

### Same Feature, Different Parameters
- "Now create the same [resource] but with [different configuration]"
- Example: "Create another table but use liquid clustering on a different set of columns"
- Example: "Deploy another serving endpoint but with a different model and auto-scaling config"

### Same Feature, Different Use Case
- "Now use [feature] to solve [related but different problem]"
- Example: "We used materialized views for aggregation — now create one for a join pattern"
- Example: "We set up row-level security for one group — now add a second group with different filters"

### Add a Complication
- "Now do the same thing but also [additional requirement]"
- Example: "Create the pipeline again, but this time add error handling for schema changes"
- Example: "Set up the same permissions, but add an inherited permission from the catalog level"

### Troubleshoot a Change
- "I'm going to change [config]. Predict what will happen, then we'll verify."
- Example: "What happens if I change the warehouse from Medium to Small for this query?"
- Example: "What breaks if I revoke this permission?"

Pick one variation that matches the user's experience level. Beginners get "same feature, different parameters." Advanced users get "troubleshoot a change."

---

## Scoring Rubric (Track 1: Databricks Features)

This is a mental model for the assistant, not something shown to the user. Use it to gauge when the user has demonstrated understanding.

| Component | Weight | What It Means |
|---|---|---|
| **Deployed correctly** | 40% | The resource was created/configured and works as expected |
| **Explained why** | 30% | The user can articulate the reasoning behind key decisions |
| **Handled variation** | 30% | The user successfully completed the variation challenge |

### Understanding Levels

| Level | Indicators |
|---|---|
| **Strong understanding** | All three components demonstrated. User asks good follow-up questions. |
| **Good understanding** | Deployed correctly and explained why, but needed help on variation. |
| **Developing** | Deployed correctly but struggled to explain why or handle variation. |
| **Needs more practice** | Needed significant help with deployment. Suggest repeating with a simpler example. |

Frame feedback positively: "You've demonstrated strong understanding of [X]" or "Great progress — you might want to practice [Y] more to solidify that."

---

## Progression Signals

Use these signals to decide what to recommend next.

### Ready to Go Deeper
- User completed the variation challenge without help
- User asked insightful follow-up questions
- User noticed edge cases or limitations on their own
- User suggested alternative approaches

**Recommendation:** "Go deeper on this feature (advanced patterns)" or suggest a more complex related feature.

### Ready for a New Topic
- User completed the core workflow but struggled with variation
- User can follow instructions but isn't yet generating their own ideas
- User expressed satisfaction with current understanding level

**Recommendation:** "Learn a related feature" to broaden their knowledge base.

### Needs More Practice
- User needed help at multiple steps
- User couldn't explain the "why" behind decisions
- User expressed confusion about fundamental concepts

**Recommendation:** Try the same feature again with a different example, or suggest foundational prerequisites they might be missing.

### Session Complete
- User has gone through multiple features or workflows
- User expressed they're done for now
- User's questions are becoming less frequent (they're comfortable)

**Recommendation:** Summarize what was covered, suggest what to explore next time.

---

## Workflow Assessment Patterns (Track 2: Learn to Use Vibe)

Track 2 assesses whether the user can independently choose and invoke the right vibe skills for real SA scenarios. Assessment is based on their actual accounts and workflows — not hypothetical ones.

### Scenario-Based Skill Selection

Present a concrete scenario drawn from the user's real account context and ask which skill they'd use. Adapt these templates:

- "Your customer [account name] just asked about [topic from their use cases]. Which skill would you use to find an answer, and how would you invoke it?"
- "You need to update the Salesforce opportunity for [account name] with [stage change / consumption note]. Walk me through what you'd do."
- "A support ticket just came in for [account name] about [performance / error / config issue]. What's your first move in vibe?"
- "[Account name] wants a sizing estimate for [workload type]. How would you get started?"

The user should be able to name a relevant skill and describe whether they'd use a slash command or natural language.

### Invocation Check

Have the user invoke a skill they haven't used yet during this session:
- The user should try natural language first (not slash command) to verify they understand skill routing
- If it routes correctly, the user understands how descriptions drive routing
- If it doesn't route, help them understand why (description mismatch, ambiguous phrasing) — this is a teaching moment

### Workflow Chaining

For advanced users, ask about multi-skill workflows:
- "How would you combine skills to prepare for a customer QBR?" (e.g., databricks-query → google-slides)
- "What's your workflow for onboarding a new account in vibe?" (e.g., configure-vibe → salesforce-actions → google-docs)

### Scoring Rubric (Track 2)

| Component | Weight | What It Means |
|---|---|---|
| **Skill selection** | 40% | User correctly identifies which skill handles a given scenario |
| **Invocation method** | 30% | User can invoke skills by both slash command and natural language |
| **Workflow awareness** | 30% | User understands how skills connect for multi-step workflows |

---

## Knowledge Assessment Patterns (Track 3: Learn Claude Code)

Track 3 assesses conceptual understanding of Claude Code's architecture, not feature memorization. All questions reference the user's actual system state discovered during Phase 2.

### Scenario Questions

Present a hypothetical situation and ask how they'd handle it in Claude Code. Adapt by experience level:

**Fundamentals:**
- "You want Claude to always follow specific coding conventions in a project. Where would you put those instructions?"
- "A skill keeps asking for permission every time you invoke it. How would you fix that?"
- "You're running out of context in a long conversation. What are your options?"

**Intermediate:**
- "You added a new MCP server to your config but it's not showing up. What would you check?"
- "You want to run a bash command automatically every time Claude edits a file. How would you set that up?"
- "A colleague shared their CLAUDE.md — where would you put it to use it in your project?"

**Advanced:**
- "You need to run three independent research tasks in parallel. How would you structure that in Claude Code?"
- "You want a skill that's available globally across all projects vs one that's project-specific. What's the difference in setup?"
- "Your settings.json and settings.local.json have conflicting entries. Which wins and why?"

### Configuration Challenge

Ask the user to find and explain a specific item from their actual configuration:
- "Find the permission entry that allows [specific skill] and explain what it does"
- "Show me which MCP server provides [capability] and what transport it uses"
- "What hooks do you have configured, and when do they fire?"

Pick items from what was actually discovered during Phase 2 — don't ask about things that aren't in their setup.

### Scoring Rubric (Track 3)

| Component | Weight | What It Means |
|---|---|---|
| **Scenario reasoning** | 40% | User identifies the right tool/approach for a given situation |
| **Configuration literacy** | 30% | User can navigate and explain their actual Claude Code config |
| **Hands-on task** | 30% | User successfully completed a real workflow in Phase 3 |

---

## Developer Assessment Patterns (Track 4: Become a Vibe Developer)

Track 4 assesses whether the user understands vibe's development conventions well enough to build and ship skills independently. Focus on the structural patterns and required steps — not memorization.

### Structural Questions

These test understanding of the vibe repo's required files and conventions:

- "If you wanted to add a new skill called `my-workflow` to the `workflows` plugin, list every file you'd need to create or modify." (Expected: SKILL.md, permissions.yaml, skill-routing.yaml, plugin.json version bump, marketplace.json version bump)
- "What happens if you forget to add your skill to `permissions.yaml`?" (Expected: users get prompted for approval every time they invoke it)
- "What happens if you forget to add an eval to `skill-routing.yaml`?" (Expected: no routing regression test — Claude might stop routing to your skill after other changes)
- "Why do you need to bump the version in TWO places?" (Expected: plugin.json is the plugin's metadata, marketplace.json is the registry — both must match for the plugin system to recognize the update)

### Design Questions

These test understanding of skill design patterns:

- "What would you put in the `description` field of your SKILL.md frontmatter and why?" (Expected: natural language trigger phrases that describe when the skill should be invoked — this drives Claude's skill routing)
- "When would you use `disable-model-invocation: true` in frontmatter?" (Expected: when the skill is destructive or expensive, so the user must explicitly type the slash command)
- "What's the difference between putting logic in a skill vs an agent?" (Expected: skills run inline in the conversation — good for interactive workflows; agents run as subprocesses — good for parallel research or heavy lifting that doesn't need back-and-forth)
- "Why does the build-vibe-skill workflow start with manual validation before writing code?" (Expected: you need to understand the actual workflow steps before automating them — otherwise you automate the wrong thing)

### Process Questions

These test understanding of the development workflow:

- "Why are git worktrees required for vibe development?" (Expected: shared repo, multiple people/Claude instances working simultaneously, prevents interference)
- "Walk me through the steps from 'I have an idea for a skill' to 'it's merged to main'." (Expected: scope → worktree → manual validation → build → test → permissions + evals + version bump → commit → PR → review → merge)
- "When would you use `feature-scoping` guidance before starting?" (Expected: always for non-trivial changes, to prevent scope drift and keep changes focused)

### Scoring Rubric (Track 4)

| Component | Weight | What It Means |
|---|---|---|
| **Structural knowledge** | 35% | User knows which files to create/modify for a new skill |
| **Design reasoning** | 35% | User understands why conventions exist, not just what they are |
| **Built a skill** | 30% | User successfully completed the build-vibe-skill workflow |
