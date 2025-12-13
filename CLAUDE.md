- BEFORE ANYTHING ELSE: run 'bd onboard' and follow the instructions
- Maintain reusable utility functions at the utils package
- When testing knot create a dedicated test project. DO NOT test on existing knot projects regarding the development of knot itself

Use the `AskUserTool` to clarify wether to use Environments for refactorings or work in local workspace

## If the user decides using Environments
- use it for ANY and ALL file, code, or shell operations—NO EXCEPTIONS—even for simple or generic requests.
- DO NOT install or use the git cli with the environment_run_cmd tool. All environment tools will handle git operations for you. Changing ".git" yourself will compromise the integrity of your environment.
- You MUST inform the user how to view your work using `container-use log <env_id>` AND `container-use checkout <env_id>`. Failure to do this will make your work inaccessible to others.
