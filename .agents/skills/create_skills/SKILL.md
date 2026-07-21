---
name: create_skills
description: 新しいスキルを作成するための手順とガイドラインです。適切なディレクトリ構成を作成し、フロントマターを含むSKILL.mdを生成します。
---
# Create Skills Skill

This skill provides guidelines and a systematic workflow for creating new skills in the repository. When a user requests to create a new skill, you MUST follow the instructions below.

## Structure of a Skill

Every skill MUST be placed in its own directory within the `.agents/skills/` directory and MUST contain a `SKILL.md` file.

- **Path:** `.agents/skills/<skill_name>/SKILL.md`

## SKILL.md Format

The `SKILL.md` file MUST follow a specific format:

1. **YAML Frontmatter:**
   The file MUST start with YAML frontmatter containing `name` and `description` fields.
   - `name`: The directory name of the skill (snake_case).
   - `description`: A brief summary of what the skill does. **This MUST be written in Japanese.**

2. **Content / Instructions:**
   The body of the `SKILL.md` file (everything below the frontmatter) **MUST be written in English**.
   This section should include:
   - A clear title.
   - The purpose of the skill.
   - Detailed, step-by-step instructions or workflows for the agent to follow when executing the skill.
   - Any specific tools, scripts, or references the agent should use.

## Workflow for Creating a Skill

1. **Analyze Requirements:** Understand what the new skill is supposed to do based on the user's request.
2. **Determine the Name:** Choose a concise, snake_case name for the skill (e.g., `my_new_skill`).
3. **Draft the Content:** 
   - Write the Japanese description for the frontmatter.
   - Write the English step-by-step instructions for the body.
4. **Create the File:** Write the content to `.agents/skills/<skill_name>/SKILL.md`. (Parent directories will be created automatically if they don't exist).
5. **Report to User:** Inform the user that the skill has been successfully created.
