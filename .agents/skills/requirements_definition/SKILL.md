---
name: requirements_definition
description: Act as a Product Manager and Systems Architect to interactively gather requirements and generate a declarative specification in the spec/ directory.
---

# Requirements Definition Skill

When the user asks for help with defining requirements or creating a specification for a new feature, you must act as an expert Product Manager and Systems Architect for the Xynon project.

## Workflow

1. **Suggest `/grill-me` (Optional but Recommended)**
   If the user's idea is still vague, suggest they use the `/grill-me` slash command. This will put you into an active interview mode to thoroughly hash out edge cases and design decisions before writing any code.

2. **Conduct the Interview**
   Ask clarifying questions to solidify the requirements. Good questions include:
   - What is the core business logic or purpose of this feature?
   - Is this an RPC plugin, a WASM plugin, or a core proxy feature?
   - What HTTP requests/responses or headers does it need to inspect or mutate?
   - Are there specific performance, security, or error-handling constraints?

3. **Draft the Specification Document**
   Once the requirements are agreed upon, draft a formal specification file inside the `spec/` directory (e.g., `spec/new_feature.md`) using your `write_to_file` tool. 

4. **Use the Standard Template**
   Always format the specification using the following Markdown template:

   ```markdown
   # Feature: [Feature Name]

   ## 1. Summary
   [Brief description of the feature's goal and business value]

   ## 2. Architecture & Component Type
   [e.g., WASM Plugin, RPC Plugin, or Core Xynon feature]

   ## 3. Behavior & Requirements
   - **Requirement 1**: [Description]
   - **Requirement 2**: [Description]
   - **Edge Cases**: [How to handle failures, timeouts, etc.]

   ## 4. Configuration
   [Example of how this feature will be defined in config.yaml]
   ```

5. **Review and Finalize**
   Present the generated specification to the user (e.g. using Artifacts) and ask if any further refinements are needed before implementation begins.
