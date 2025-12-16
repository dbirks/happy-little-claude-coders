# Implementing Multi-Repo Workspaces with Claude Code, Happy CLI, and Automated Releases

## Transition from “Instances” to **Workspaces**

The development environments will be conceptualized as **workspaces**, aligning with terminology used by Anthropic and others. A workspace is essentially an isolated dev environment that can hold one or multiple project repositories and associated tools. Anthropic’s console, for example, uses “Workspaces” to mean unique environments for organizing resources and managing multiple deployments[[1]](https://claude.com/blog/workspaces#:~:text=We%27re%20introducing%20Workspaces%20in%20the,on%20a%20more%20granular%20level). In our context, each workspace will provide an isolated containerized environment (with its own storage and config) where an AI coding agent (Claude Code, referred to internally as *CLAWD Code*) can operate. This change is mainly semantic but important for clarity – instead of saying “spin up an instance”, we’ll say “launch a workspace.” Workspaces make it clear that each environment is a self-contained sandbox (with its own repos, tools, etc.) that can be created and destroyed independently.

Key characteristics of our **workspaces**:

* **Isolated Environment:** Each workspace runs in its own container (on Kubernetes via our Helm chart) with dedicated resources and volumes. This isolation ensures the agent’s actions (file edits, commands) are sandboxed.
* **Multiple Repositories:** A single workspace can contain *multiple Git repositories* simultaneously. This is useful for multi-repo projects (monorepos or related services). We’ll configure the workspace to clone a list of repos on startup (see below), instead of being tied to a single repo.
* **User-Friendly Context:** Workspaces will be set up so that if only one repo is present, the agent/user is placed directly in that repository’s directory. If multiple repos are present, the workspace’s working directory will be a parent folder containing all repo folders, allowing the agent to see all project code at once. This approach ensures the AI (or developer) can easily navigate and reference all necessary code.

*(By adopting “workspace” terminology, we emphasize the environment’s role in organizing resources, similar to how Anthropic’s Workspaces help group API keys, usage, etc., per environment*[*[2]*](https://claude.com/blog/workspaces#:~:text=For%20developers%20using%20Claude%20across,organization%20and%20individual%20API%20keys)*. Here it groups code and config for an AI agent.)*

## Base Container Image and Essential Tools

We will build the Docker container for the workspace on a **Debian Linux base** (e.g. Debian 12 “Bookworm” slim). A Debian base provides a stable and minimal foundation, with wide compatibility for installing development tools. On top of this, we need to install several tools and dependencies that our AI coding agent and the “Happy” CLI require:

* **Git** – for version control operations. We’ll ensure the latest Git client is installed so the agent can clone repositories, create branches, commit and push changes, etc.
* **GitHub CLI (gh)** – this is crucial for authenticating with GitHub and performing GitHub-specific actions. The GH CLI lets us script logins, repository cloning, PR creation, issue management, etc., right from the terminal. We will install it via the official package or apt repo. (The GH CLI respects a GITHUB\_TOKEN env var for authentication if set[[3]](https://www.boxpiper.com/posts/github-cli#:~:text=,hostname%20%3Chostname), which will be useful for non-interactive bot login as discussed later.) Using the GH CLI provides a convenient way for the agent to interact with GitHub programmatically beyond basic git. For example, after logging in, gh repo clone can clone a repo with proper credentials, gh pr create can open pull requests, etc.
* **Node.js 20+** – Node is required to run the **Happy CLI** (which is written in TypeScript). The Happy CLI documentation states Node.js ≥ 20.0.0 is needed[[4]](https://github.com/slopus/happy-cli#:~:text=Requirements). We will likely use Node 20 LTS. We can base our image off an official Node 20 image (which in turn is Debian-based) or install Node via the NodeSource repository. Ensuring Node is available means we can run or even develop the Happy CLI within the workspace.
* **cURL and other utilities** – We’ll include cURL (needed to install some tools like Claude CLI via script), and typical developer utilities (possibly bash, coreutils, etc., though Debian base includes most). Since the container might be used interactively by an agent, having conveniences like htop or editors isn’t strictly necessary for the AI, but minimalism is okay unless needed.

**Claude Code CLI (CLAWD Code)**: We will install the Claude Code CLI tool (Anthropic’s CLI for their coding assistant) inside the container. Anthropic provides an easy install script:

curl -fsSL https://claude.ai/install.sh | bash

Running this will download and install the latest Claude CLI binary, making the claude command available globally[[5]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=To%20install%20Claude%20Code%2C%20log,and%20run%20the%20following%20command). After installation, one can verify with claude --version[[6]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=curl%20). The Claude CLI allows our agent to run, providing the interface for Claude to act on the codebase. We will **not** be enabling Claude’s “teleport” feature in this environment – the teleport flag (which allows moving a session between devices) is unnecessary here, since our use-case is a single-agent container. Disabling it simplifies configuration. (Teleport “allows you to start a task from one device and pick it up on another”[[7]](https://shipyard.build/blog/claude-code-on-the-web/#:~:text=CC%20web%20also%20has%20a,when%20you%20might%20be%20AFK), but in a cloud workspace context we won’t be switching devices, so we can omit any --teleport usage.)

**Happy CLI**: The Happy CLI (“Happy Coder”) is the tool that bridges Claude Code with mobile/web clients, allowing session streaming. We have two options to get Happy CLI in the workspace: 1. **Install via npm** – e.g. npm install -g happy-coder. The Happy CLI README indicates it can be installed globally via npm[[8]](https://github.com/slopus/happy-cli#:~:text=Installation). This will put a happy binary in our PATH. Using npm ensures we get a released version easily. 2. **Clone from source** – We also intend to clone the happy-cli GitHub repository into the workspace (so the agent can read/modify it). The default repo list (discussed below) will include Happy CLI’s source. Even if we install it globally for usage, having the source code present means the AI agent could potentially update or refer to it. The Happy CLI project is open-source (MIT) and has an active repo at slopus/happy-cli[[9]](https://github.com/slopus/happy-cli#:~:text=device%20github,cli).

Given that Happy CLI requires Claude CLI to actually function (it launches Claude Code and monitors it[[10]](https://github.com/slopus/happy-cli#:~:text=This%20will%3A)), we will ensure: - Claude CLI is installed and **authenticated** (see next section). - Happy CLI is installed and/or its code is present.

Finally, we might include any other helpful tools: for example, if we want the agent to have Docker or kubectl or other CLIs, we could include them. But as of now, the critical ones are those above. In summary, **our Dockerfile** will start from a Debian base, then install Git, Node 20, and GH CLI (via apt or official script), and use npm to install Happy CLI. It will also run the Claude CLI install script. After these steps, the container will have: - git – ready to use. - gh – ready (but not yet authenticated until runtime). - node and npm – with happy CLI installed globally. - claude – the Claude Code CLI, which at runtime will need login.

## Authentication Setup: GitHub Bot Login & Claude API Access

**GitHub Bot Authentication:** Since the AI agent will need to pull and push code from GitHub (and possibly open PRs), we will set up a **GitHub bot account** (or use a PAT from a service account) for authentication. The GitHub CLI gives us a few ways to authenticate non-interactively: - **Environment Token:** We can provide a Personal Access Token (PAT) as an environment variable GITHUB\_TOKEN or GH\_TOKEN. The GH CLI will automatically respect GITHUB\_TOKEN if set[[3]](https://www.boxpiper.com/posts/github-cli#:~:text=,hostname%20%3Chostname). This means if our Kubernetes deployment injects a secret PAT into the container environment, the gh tool will pick it up and we can immediately use gh (or even plain git with HTTPS) without manual login prompts. - **Scripted Login:** Alternatively, we could run gh auth login --with-token during container start, feeding in the token. For example: echo "$GH\_PAT" | gh auth login --with-token. The GH CLI documentation shows that gh auth login can read a token from stdin and authenticate[[11]](https://www.boxpiper.com/posts/github-cli#:~:text=,setup). This would store the credentials in GH’s config file.

We will likely use the environment token method for simplicity (the token will come from a Kubernetes Secret for security). By doing so, the first time the container’s entrypoint runs, the GH CLI will see the token and create an authentication entry for GitHub. In fact, we can combine this with configuring Git itself. If gh auth login is run with the --git option or the user chooses to “Authenticate Git with GitHub credentials”, GH CLI will set up git credential helpers so that git clone over HTTPS uses the stored token[[12]](https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode#:~:text=The%20solution%20is%20simple%3A)[[13]](https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode#:~:text=1.%20Run%20,time%20code%204.%20Done). We want to script this, so we don’t need any interactive prompt. Using GH\_TOKEN env means even commands like gh repo clone or git push will just work.

**Persistent Credentials Storage:** To avoid having to re-authenticate the bot on every workspace start, we will **persist the GH CLI config directory** on a volume. By default, GH CLI stores auth info in ~/.config/gh/hosts.yml (with tokens for each host) and settings in ~/.config/gh/config.yml. We will mount a PersistentVolumeClaim (PVC) at this path (or at least mount ~/.config/gh onto a PVC). This way, once the bot logs in the first time, the credentials are saved on the persistent volume. Subsequent restarts of the pod or new pods for that same workspace can reuse the token from the volume. In other words, the PVC acts as the bot’s “home directory” for credentials.

Persisting data across container restarts is crucial because container filesystems are otherwise ephemeral – *“whenever you write anything to the container filesystem, it will be removed when the container restarts”*[[14]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=Generally%2C%20containers%20are%20stateless,removed%20when%20the%20container%20restarts). By using a persistent volume, we ensure the auth data **“survives” container restarts, like plugging a USB stick into each new container**[[15]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=Persistent%20volumes%20allow%20you%20to,hard%20drive%20on%20every%20restart). Only the bot’s config (which is small) will live on this volume.

Implementation details for the Helm chart: - We will add a **PVC template** for the GH config. For example, define a volume gh-config with a default size (tiny, e.g. 1Gi or even less) and mount it at /root/.config/gh (assuming the container runs as root user – more on that below). The user can override the size or disable persistence if desired via values. But by default, this will be enabled to avoid repeated logins. - Alternatively, if running as non-root user (for security), we’d mount at /home/<user>/.config/gh. The user in the container could be root since it’s a dev agent environment (the prompt log suggests running as root is okay). We’ll proceed assuming root for simplicity. - The **bot’s token** will be provided as a Kubernetes Secret (mounted as env var GH\_TOKEN). On container start, an entrypoint script can check if ~/.config/gh/hosts.yml exists; if not, run gh auth login --hostname github.com --with-token <<<"$GH\_TOKEN". Once done, the token will be saved to hosts.yml on the PVC. (As a note, one can revoke local copies of credentials by removing that hosts.yml[[16]](https://developer.1password.com/docs/cli/shell-plugins/github/#:~:text=Use%201Password%20to%20authenticate%20the,config%2Fgh%2Fhosts.yml.%20Next%20steps), but here we want to keep it).

Additionally, we should configure Git’s identity for the bot so commits have a proper author. We can bake in something like:

git config --global user.name "YourBotName"
git config --global user.email "yourbot@users.noreply.github.com"

This could be done in the Dockerfile or at runtime via config map. This ensures any git commit the agent does will have an identifiable signature.

**Claude CLI Authentication:** The Claude Code CLI will also require authentication – it needs access to an Anthropic Claude API or account. Typically, running claude /login opens a browser login flow[[17]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=Logging%20in). In a headless container, that’s tricky. However, Anthropic might allow using API keys or a headless login for Claude Code. Possible approaches: - If Claude Code supports using an API key, we could set an env var or config file for it (Anthropic’s docs mention “workspace-scoped API keys” in console, but for CLI it’s not clear). If not, manual login might be needed initially. One solution is to use the **Teleport** feature we decided to skip: originally, you could start a session locally and “teleport” it to the cloud. But since we’re not doing teleport, an alternative is needed. - We might have to run claude /login once manually per workspace to authenticate (which requires copying a URL and logging in). This is not ideal for automation. If Anthropic’s CLI allows environment variables for auth (some CLIs allow setting CLAUDE\_API\_KEY), we’ll use that. Otherwise, perhaps the Happy CLI (which starts Claude) could pass in API tokens. The Happy CLI has commands like happy connect to store API keys in the Happy cloud and happy auth[[18]](https://github.com/slopus/happy-cli#:~:text=Commands) – possibly that’s how it manages Claude API credentials. In *Happy’s case, the README indicates you must have Claude CLI “installed & logged in” before starting Happy*[[19]](https://github.com/slopus/happy-cli#:~:text=,command%20available%20in%20PATH). - One possible path: Use **Anthropic API Key** if available. The Claude CLI might accept an API key config (some users mention npm install -g @anthropic-ai/claude-code for local usage which might prompt for API key or login).

For now, we’ll assume an initial manual step to authenticate Claude CLI (copy link, etc.) which will save credentials (likely under ~/.config/claude or similar). If that’s the case, we could also persist Claude’s config on the same or another PVC if needed. However, since our agent is mostly autonomous, we might prefer to find a programmatic way. This detail may require further research or using an alternate approach such as orchestrating Claude Code via their SDK.

*(We can iterate on this, but it’s somewhat orthogonal to the main dev environment setup. The key point is GH auth is handled via token and persistent storage, whereas Claude auth needs either manual login or an API key approach.)*

## Helm Chart Configuration: Repository Cloning

One of the core features will be the ability for a workspace to automatically **clone a list of Git repositories** when it starts. We will introduce a new Helm chart value (e.g., workspace.repos) which is an array of Git repo URLs (or GitHub org/repo identifiers) to clone. The user (or automation) can populate this list per workspace.

**Default Repos:** If no list is provided, we will default to cloning the **Happy CLI repository** by default. This ensures that a fresh workspace always has at least the happy-cli code available (since our agent’s focus is on that project). The default can be something like:

workspace:
 repos:
 - "https://github.com/slopus/happy-cli.git"

Users can override this repos list via their values (e.g., to add other project repos or replace the default).

**Cloning Strategy:** We have a few ways to perform the git clones on pod startup: - **Init Container:** Use a lightweight init container (with git installed) that runs before the main container. The init container can pull the code and deposit it into a shared volume mount. This approach cleanly separates cloning logic and ensures the code is ready before the main process (Claude/Happy) starts. For example, an init container using the alpine/git image can do something like:

initContainers:
 - name: clone-repos
 image: alpine/git:latest
 args:
 - /bin/sh
 - -c
 - |
 git clone --single-branch --depth 1 "$REPO\_URL" /workspace/repo1 && \
 git clone ... (for others)
 volumeMounts:
 - name: workspace-code
 mountPath: /workspace
 env:
 - name: GIT\_USERNAME
 valueFrom: ... (if using basic auth)
 - name: GIT\_PASSWORD
 valueFrom: ... (if using basic auth)

The above is pseudo-code; a real example from a tutorial shows an initContainer using an Alpine Git image to clone a repo using credentials from a Kubernetes Secret[[20]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=initContainers%3A%20,%27https%3A%2F%2F%24%28GIT_USERNAME%29%3A%24%28GIT_PASSWORD%29%40gitlab.company.com%3E%2Fpath%2Fto%2Frepo.git)[[21]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=env%3A%20,GIT_PASSWORD%20valueFrom%3A%20secretKeyRef%3A%20key%3A%20password). In that example, the init container took the approach of embedding the username/password into the clone URL (i.e., https://$(GIT\_USERNAME):$(GIT\_PASSWORD)@gitlab.company.com/path/to/repo.git)[[22]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=,%27https%3A%2F%2F%24%28GIT_USERNAME%29%3A%24%28GIT_PASSWORD%29%40gitlab.company.com%3E%2Fpath%2Fto%2Frepo.git). We can do something similar if needed for private repos. However, since we have GH CLI auth set up, we might not need to pass raw credentials – the main container can handle clones. - **Main Container Script:** Alternatively, we could have the main container’s startup script perform the clones (before launching Claude Code). For instance, the entrypoint could iterate over the list of repos and run gh repo clone ... or git clone .... Because we will have the GH CLI authenticated in the main container, using gh repo clone owner/name is convenient (GH CLI will use the stored token). Or we can do git clone https://github.com/owner/repo.git – with the GH credential helper set up, Git will also use the token. This approach avoids needing a separate init container, but it means our entrypoint script becomes a bit more complex (especially handling multiple clones sequentially).

Using an **init container** is a clean solution, especially if some repos might be large – the init can run to completion before the main process starts. It also allows allocating separate resource limits for the clone process if needed. We will likely go with an initContainer for cloning. We’ll create a shared volume (an emptyDir) named, say, workspace-code, mounted at /workspace in both init and main container. The init container will run a small script to clone all repos listed.

For private repos: if the GH CLI authentication is done in the main container, the init (which is separate) won’t have that config. We have two choices: - Pass a PAT via environment into the init and use it in the git URLs (as shown above)[[21]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=env%3A%20,GIT_PASSWORD%20valueFrom%3A%20secretKeyRef%3A%20key%3A%20password). This means duplicating the secret in init env. It’s secure if handled properly (env var from K8s secret is fine). - Alternatively, forego init for private repos and let main do clones (so it can rely on GH CLI). But mixing approaches could be messy.

Perhaps simpler: use the PAT in init clone for all repos (public or private). Public repos can be cloned without auth anyway, so that’s fine. For uniformity, we might always supply GIT\_TOKEN (our PAT) and do git clone https://$GIT\_TOKEN@github.com/owner/repo.git. This avoids interactive prompts and doesn’t require GH CLI in init. Storing a PAT in the pod is okay (it’s the same token GH CLI uses). The init container method in the Medium article essentially did that via username/password secret[[23]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=valueFrom%3A%20secretKeyRef%3A%20key%3A%20username%20name%3A,secret).

**Volume and Path Structure:** We will mount an **emptyDir** volume at /workspace (inside the container). All repositories will be cloned into this directory. Here’s how we’ll organize it: - If only **one repository** is in the list, we will clone it *directly into /workspace*. This can be achieved by specifying the target directory in the git clone command. For example: git clone https://github.com/owner/repo.git /workspace. This way, the repository’s files populate the /workspace directory (with its .git folder there). In this scenario, **the user/agent’s working directory will be /workspace**, which is the root of that repo. This meets the requirement “if there's only one repo, start the user straight in that repo” – effectively, they don’t have to cd into a subfolder. - If **multiple repositories** are listed, we will clone each into its own subdirectory under /workspace. By default, git clone without a target will create a folder named after the repo. For example, cloning happy-cli.git would make /workspace/happy-cli/. We can also be explicit: git clone URL /workspace/happy-cli. We will do this for each. In this case, the **working directory will remain /workspace** itself (containing subfolders for each repo). This way, the agent can easily see all repo directories. The Happy CLI expects, for instance, that if multiple repos are present, Claude Code can handle them (Claude Code will index the entire /workspace directory). Starting in the parent directory ensures the agent isn’t “inside” one repo and blind to the others – it can navigate between them.

We will implement logic (likely in a script) to detect the number of repos:

REPO\_COUNT=${#REPO\_LIST[@]}
if [ "$REPO\_COUNT" -eq 1 ]; then
 # Single repo: clone into /workspace directly
 git clone "${REPO\_LIST[0]}" /workspace
 # Optionally, cd into it (not needed if entrypoint does exec from correct dir)
else
 # Multiple repos: clone each into its own subdir
 for url in "${REPO\_LIST[@]}"; do
 repo\_name=$(basename -s .git "$url") # derive name from URL
 git clone "$url" "/workspace/$repo\_name"
 done
 # Working dir stays as /workspace
fi

The above is conceptual – actual implementation might differ slightly in Helm templating or an entrypoint script file. We might ship an entrypoint script with the image to handle this logic, reading an env like REPOS="repo1 repo2 ..." passed by the chart.

**Starting Directory:** In Kubernetes Pod spec, we can set the container’s workingDir: "/workspace" so that when the main process starts, it’s already in /workspace. For a single repo scenario, /workspace is the repo root (as we cloned into it). For multi-repo, /workspace is the parent (with repos inside). In either case, this achieves the desired behavior regarding starting location. If we wanted to be fancy, we could programmatically cd in the entrypoint for the single-repo case, but simply cloning into the correct path is cleaner.

By organizing the code like this, **Claude Code** (CLAWD) and **Happy CLI** will see the appropriate files: - Claude Code typically indexes the current directory’s files for context. If one repo, that’s the repo’s files; if many, having them in subfolders is fine since Claude can still open them if instructed, or we can possibly configure it to index recursively. - The Happy CLI uses Claude under the hood and doesn’t impose restrictions on working dir beyond requiring Claude CLI available. So it should be fine.

## Storage Configuration: Ephemeral vs Persistent Volumes

We touched on persistence for GH auth. Now let’s formalize the storage setup in the Helm chart for both code and config:

* **Code Volume (Workspace storage):** We will use an **emptyDir volume** for the code under /workspace. An emptyDir is a transient volume that lives as long as the pod lives and is destroyed when the pod is removed. This is suitable for ephemeral development environments. We don’t necessarily want to persist the code beyond the life of the workspace because any changes should be pushed to GitHub anyway, and if the workspace is re-created, it can just clone fresh. Using emptyDir also avoids leaving possibly large data around on cluster storage after the workspace is gone. We will, however, make the size of this volume configurable. Kubernetes emptyDir supports an optional sizeLimit field to constrain how much data can be stored[[24]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=emptyDir%3A%20%20%20%20,Kubernetes%20EmptyDirVolumeSource%20v1). For example, we might default to 10Gi, which is plenty for most repositories and some build artifacts:
* volumes:
   - name: workspace-code
   emptyDir:
   sizeLimit: 10Gi # default, can override via values
* In values.yaml, we’ll add something like workspace.storageSize: 10Gi so users can override it (e.g., set 20Gi for bigger monorepos). This gives per-workspace control over storage. The phrase “same defaults, just with overrides” indicates we won’t change the default size that was previously used (if any), just make it configurable.

If a user did want to persist code (perhaps to retain unpushed changes between restarts), they could modify the chart to use a PVC for /workspace. By default, we’ll use ephemeral storage, but our design should be flexible if persistence is later desired (maybe via a flag). Persisting code isn’t usually necessary for ephemeral dev envs, and can even be undesirable if multiple parallel workspaces use the same PVC, so we stick to emptyDir by default.

* **Persistent Config Volume:** As discussed, we will define a PVC for GH CLI config (and possibly Claude CLI config if needed). This PVC can be relatively small (even 100Mi is enough for some text config, but 1Gi as a safe minimum). In values.yaml we can have:
* githubAuth:
   enabled: true
   pvcSize: 1Gi
   storageClass: default
* If enabled, the deployment will include a PVC volumeMount at ~/.config/gh. We might not expose storageClass in values for simplicity (it can just use cluster default unless specified). The key is that this volume persists even if the workspace pod is deleted, so that a new pod (with the same claim) can reuse it. This is especially relevant if our workspace is tied to a longer-lived concept (like a user’s persistent environment). However, if our workspaces are truly ephemeral per PR for example, we might not reuse the same PVC for new pods (since each PR env might be separate). In a scenario where each PR gets a new workspace (with its own fresh PVC), the persistence might be moot unless the pod restarts within that PR’s life.

Another approach is to mount a **Kubernetes Secret or ConfigMap** with the token instead – but that wouldn’t allow storing new data like a generated hosts.yml. So PVC is the way to go for dynamic storage.

**User Home and Permissions:** We should confirm the user under which the container runs. Running as root simplifies things (no permission issues writing to /workspace or /root/.config). We can run as a non-root user for security; if so, we’d adjust mount paths (/home/<user>/... and ensure volumes permission via fsGroup, etc.). Since this is an internal dev tool, we might accept root user for expediency (the user said “Root, uh, yeah. Debian has a base. yeah, git, gh.” implying root is fine).

To sum up, the **Helm chart’s volume setup** will look like:

volumeMounts:
- name: workspace-code
 mountPath: /workspace

- name: gh-config
 mountPath: /root/.config/gh # only if GitHub auth persistence is enabled

volumes:
- name: workspace-code
 emptyDir:
 sizeLimit: {{ .Values.workspace.storageSize | default "10Gi" }}

- name: gh-config
 persistentVolumeClaim:
 claimName: {{ include "mychart.fullname" . }}-gh-config # PVC defined elsewhere

And we’d have a PVC template for gh-config using the size from values.

With this, each workspace gets the same defaults (e.g. 10Gi ephemeral storage) unless overridden, and an optional persistent volume for GH creds.

## Running the Happy & Claude CLI in the Workspace

Once the repositories are cloned and volumes mounted, the final piece is to **launch the AI agent process**. This likely means running the *Happy CLI*, which in turn starts a Claude Code session.

The container’s entrypoint command could be something like: happy (which starts the CLI and thereby Claude). However, we might want to pass some flags: - Happy CLI might need to know which model to use (-m sonnet by default) – we can let it default. - It might need the Claude CLI’s path (it usually just calls claude internally if available in $PATH). - We might want to run happy daemon or some command to keep it running in the background. But happy on its own likely runs an interactive session. Possibly, for an unattended agent, we might prefer running Claude CLI directly. But using Happy allows the mobile/web integration and push notifications – which may or may not be needed for an automated agent.

Given the context, it sounds like the goal is to allow a *future coding agent (AI)* to pick up this environment. Possibly, the agent controlling logic might not rely on Happy’s mobile sync features. It might be sufficient to just run Claude CLI in auto-mode. However, the user specifically mentions Happy CLI, so likely they want it in the loop (maybe to enable connecting a mobile UI to monitor progress).

We can decide to launch happy by default. The **Helm chart** can have something like:

args: ["happy", "--claude-arg", "--dangerously-skip-confirm"]

(for example, passing any needed Claude flags via Happy’s --claude-arg). Or if we don’t want to use Happy, we could launch Claude directly with certain flags. Since the user emphasized Happy CLI, we’ll assume we run that. The good thing is, Happy will handle ensuring Claude is started and it will provide a QR code / link for connection. This is fine if a human is connecting; for pure automation, it might not matter.

Regardless, by the time happy (or claude) runs, the working directory will have the code repositories, and the GH auth will be configured. The agent will be able to read and modify files and use git. We may consider pre-seeding some context or instructions to Claude about the repos (Claude Code sometimes reads a claude.md file for instructions). For instance, we could drop a CLAUDE.md file at the repo root describing project context or objectives. That could be a nice-to-have: e.g., a brief description of each cloned repo, to help the AI understand what it’s working with. But that’s outside the immediate infra setup; it’s more about prompt engineering.

## Automated Versioning and Release Management

We want to implement automated version bumping and releasing for both the Docker image and the Helm chart using **Release Please**. This will ensure that changes to the project (especially conventional commits) trigger appropriate version increments, changelogs, and releases of our artifacts (container image & chart).

**Conventional Commits:** We will adopt the **Conventional Commits** style for commit messages (if not already). This means commit messages should be prefixed with types like fix: ..., feat: ..., chore: ..., etc. These prefixes will determine the next version number. Release Please specifically looks for these to decide how to bump versions: - A commit prefix fix: indicates a bug fix and results in a **patch** version bump (X.Y.Z -> X.Y.(Z+1))[[25]](https://github.com/googleapis/release-please#:~:text=The%20most%20important%20prefixes%20you,should%20have%20in%20mind%20are). - A commit prefix feat: indicates a new feature and results in a **minor** version bump (X.Y.Z -> X.(Y+1).0)[[25]](https://github.com/googleapis/release-please#:~:text=The%20most%20important%20prefixes%20you,should%20have%20in%20mind%20are). - If the commit message has a **bang** ! (e.g. feat!: ... or fix!: ...) that signals a **breaking change**, which triggers a **major** version increment[[26]](https://github.com/googleapis/release-please#:~:text=,result%20in%20a%20SemVer%20major). - Other types (docs, chore, refactor without !) by default do not bump version, unless configured otherwise. We might treat them as patch or ignore them in the release notes.

Using this convention means developers (or AI agents committing) need to format commit messages accordingly. It might be worth adding a note to the contributors that this style is required for automated releases.

**Release-Please GitHub Action:** We will utilize the **release-please GitHub Action** in our repository. The typical setup is: - A configuration file (either in release-please-config.json or in GitHub Actions workflow yaml) specifying how to handle releases. - We likely will use the **manifest** mode of release-please, since we have multiple artifacts: at least the npm-like project (even if not published) and the Helm chart (and possibly the Docker image).

However, we can simplify by treating the whole repository as a single “package” and use **extra files** feature to update Chart.yaml and others. Our plan: - Keep a top-level package.json to track the overall version of the project (for “CLAWD code and Happy CLI”). Even if we don’t publish to npm, having a version in package.json gives release-please a natural place to bump and also is a central version for the Docker image. (Release Please automatically updates package.json version if it finds one[[27]](https://github.com/googleapis/release-please#:~:text=1,Release%20based%20on%20the%20tag), which we want). - Mark the relevant lines in Chart.yaml and perhaps values.yaml with the release-please annotation comment so they get bumped too. According to release-please docs, you can add # x-release-please-version next to the version lines in arbitrary files and release-please will update those lines[[28]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=,version). We will do this for: - Chart.yaml: the version: field, and possibly the appVersion: field if we want to keep it in sync with the app’s version. - values.yaml: the image tag value if one exists (e.g., if our Helm chart has image.tag: v1.2.3 we put the comment there). - We might also have a README.md or other references, but the above are key.

For example, in Chart.yaml:
```yaml
version: 0.1.0 # x-release-please-version
appVersion: "0.1.0" # x-release-please-version
```
And in values.yaml:
```yaml
image:
 repository: myrepo/clawd-workspace
 tag: "0.1.0" # x-release-please-version
```
This tells release-please to bump those strings. A similar approach is seen in community projects, e.g., a Helm chart’s appVersion line annotated for release-please【26†L15-L23】.

* We’ll configure release-please in **manifest mode** so it can handle multiple files. The manifest config (often .release-please-manifest.json) can list packages. But since we treat it mostly as one package, we might get away with a simpler config. Possibly:
* {
   "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
   "release-type": "node",
   "packages": {
   ".": {
   "release-type": "node",
   "extra-files": ["Chart.yaml", "helm/ourchart/Chart.yaml", "helm/ourchart/values.yaml"]
   }
   }
  }
* (If our chart is in a subdirectory, we specify the path; if it’s in root, then just Chart.yaml is fine.)
* Alternatively, since Chart.yaml isn’t a typical version file, we rely on the annotation rather than extra-files in config. The blog by Mark Cheret demonstrates using extra-files + annotation for a non-standard file[[29]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=%7B%20,false)[[30]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=The%20file%20gets%20its%20version,Add%20the%20annotation). We’ll do similarly.

**Release PR flow:** With release-please configured, here’s how the workflow will go: - Developers (or the AI agent) merge feature/fix branches into main with conventional commit messages. - **Release-Please Action** (running on each push to main) will analyze the commit history since the last release tag. If it finds new commits with feat/fix, it will open a **Release PR**. This PR will be titled like chore(main): release 0.2.0 (for example) and contain: - An updated CHANGELOG.md with entries generated from the commit messages (grouped into Features, Bug Fixes, etc., based on commit type). - Bumped versions in package.json, Chart.yaml, etc. (all files with the annotation) to the new version number. - Possibly an updated Git tag in some manifest (but mainly it prepares for tagging).

*“Whenever a pull request is merged to main, release-please will open a release pull request and take care of bumping the version in the Chart.yaml file”*[[31]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=Helm%20Charts) – this matches our use: when our main branch changes with relevant commits, it triggers the release-please workflow to create a PR that bumps Chart.yaml’s version (and others).

* We (or an automated process) review that PR. It will show the proposed new version and changelog. The version bump follows semver rules: e.g. if we had feat: commits, minor is bumped, etc., as per Conventional Commits[[32]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=3.%20Release,the%20packaged%20helm%20charts%20are). For example, *“fix: port in service will bump the patch version, feat: add new service will bump the minor version”*[[33]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=4,the%20packaged%20helm%20charts%20are). Breaking changes (feat! or fix!) would have bumped major.
* When ready, we **merge the Release PR** into main. Upon merge, the release-please action will:
* Tag the commit on main with the new version (e.g., v0.2.0).
* Create a GitHub Release on our repository, with the changelog and version.
* (If configured) mark the PR as released.

This process is described in release-please’s documentation: *on merging the release PR, it updates changelog and version files, tags the commit with the version, and creates a GitHub release*[[34]](https://github.com/googleapis/release-please#:~:text=When%20the%20Release%20PR%20is,please%20takes%20the%20following%20steps). We should see labels like autorelease: pending on the PR initially and autorelease: tagged once it’s merged and tagged[[35]](https://github.com/googleapis/release-please#:~:text=,a%20convention%20for%20publication%20tooling).

* Because we want to **version the Docker image** as well, we will tie the Docker image tag to this release tag. Specifically, we will set up a GitHub Actions workflow that triggers on the creation of a new Git tag (or on release events). When a new version tag (say v0.2.0) is pushed by release-please, our action will:
* Check out the repo at that tag.
* Build the Docker image (the workspace container) using the Dockerfile.
* Tag the image with 0.2.0 (and maybe also latest).
* Push the image to our registry (could be Docker Hub or GitHub Container Registry).

In the wundergraph/cosmo project, for example, when Lerna updated versions and tags, they had a workflow to build and push images with tags like 0.4.2, latest, and a commit SHA[[36]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=GitHub%20container%20registry%2C%20Lerna%20will,0.4.2). We can follow a similar pattern. *“The workflow will build and tag all images: latest (for main branch), sha-<short> for caching, and git-tag e.g. 0.4.2”*[[37]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=tag%20all%20images%3A%20,be%20overwritten%20on%20every%20release). For us, tagging with the semantic version is most important.

We will likely not use Lerna (that was for a monorepo with multiple packages), but our GitHub Actions can directly call docker build. For example, a workflow file .github/workflows/release.yml could:

on:
 push:
 tags:
 - 'v\*.\*.\*'
jobs:
 build-and-push:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v3
 - name: Set up Docker Buildx
 uses: docker/setup-buildx-action@v2
 - name: Login to GHCR
 uses: docker/login-action@v2
 with:
 registry: ghcr.io
 username: ${{ github.repository\_owner }}
 password: ${{ secrets.GHCR\_TOKEN }}
 - name: Build image
 run: |
 docker build -t ghcr.io/our-org/our-image:${GITHUB\_REF#refs/tags/} .
 - name: Push image
 run: |
 docker push ghcr.io/our-org/our-image:${GITHUB\_REF#refs/tags/}
 # Also push alias 'latest' if needed
 docker tag ghcr.io/our-org/our-image:${GITHUB\_REF#refs/tags/} ghcr.io/our-org/our-image:latest
 docker push ghcr.io/our-org/our-image:latest

This ensures each release produces a versioned Docker image. (If using Docker Hub, just change the registry and login accordingly.)

* For the **Helm chart release**, we should also package and publish the chart when we cut a release. Since our Chart.yaml version will be updated by the release PR, once that PR merges, the main branch has the new Chart version. We can automate chart publishing in a couple of ways:
* **Publish to a Helm repo (e.g., GitHub Pages or OCI):** We can use GH Actions to package the chart (helm package . to get a .tgz) and then push it. A modern approach is to push Helm charts to an OCI registry (Helm supports OCI artifact repositories). In fact, the Cosmo example pushes charts to an OCI registry on GHCR[[38]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=6,router). *“As soon as the PR gets merged, the packaged helm charts are pushed to oci://ghcr.io/... After some time the released version will be available on Artifact Hub”*[[38]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=6,router).
* Alternatively, if we maintain a GitHub Pages branch for charts, we could commit the packaged chart there.

A simpler route: publish chart as OCI to GHCR alongside the Docker image. We already have GHCR auth in the workflow from above. We can add steps:

- name: Install Helm
 uses: azure/setup-helm@v3
- name: Push Helm Chart to OCI
 run: |
 helm package ./helm/ourchart -u -d ./ # package into .tgz
 helm registry login ghcr.io -u $USER -p $PAT
 helm push ourchart-0.2.0.tgz oci://ghcr.io/our-org/charts

This will push the chart archive to GHCR (where it can be fetched with helm pull oci://ghcr.io/our-org/charts/ourchart --version 0.2.0). We would do this on tag as well.

If publishing to Artifact Hub is a goal, using OCI and then listing that on Artifact Hub is a common approach. We should increment the chart version exactly with app version to avoid confusion. If we ever need the chart version to differ (e.g., chart changes without app change), that’s an edge case; but since our chart primarily deploys this app, keeping them in lockstep is fine (chart version == app version).

**Release Please Config for Multi-Components:** If we wanted to manage versions of multiple components independently (say, if Happy CLI and Clawd Code were separate), Release Please supports a **monorepo configuration**. But in practice, we’re treating everything as one unit/version, so independent versioning isn’t needed. (As an aside, a blog noted release-please can’t independently version a single file or sub-component without a directory; the workaround was to use extra-files so that the file shares the main version[[39]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=The%20configuration%20path%20must%20point,a%20directory%20containing%20versionable%20files)[[40]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=annotation%3A). We’ve applied that workaround by treating Chart.yaml’s version as tied to the main version.)

**Changelog and Documentation:** With release-please, we’ll get a nicely maintained CHANGELOG.md that accumulates changes. Each release PR will append a section for the new version with each commit message summarized. This is great for tracking what the AI agent did between versions as well. We should ensure our commit messages are clear and maybe use the body of commit message for more details if needed – those get into the changelog too.

We should also consider adding badges or status in the README for version, etc., but that’s minor.

Finally, we will integrate **Release Please into our CI** so that it runs automatically. Likely: - Use the official google-github-actions/release-please-action. For example, in .github/workflows/release-please.yml:

on:
 push:
 branches: [ main ]
jobs:
 release:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v3
 - uses: google-github-actions/release-please-action@v3
 with:
 config-file: .release-please-config.json
 token: ${{ secrets.GITHUB\_TOKEN }}

This will cause the Release PR to be opened when needed. We might refine conditions (perhaps only on main and when conventional commits are present).

* We ensure our config file specifies the correct release type. “node” release-type is fine since it knows how to bump package.json. We may need to specify bump-minor-pre-major: true if we want feat to bump minor even <1.0 (though by default, I think release-please might bump minor anyway). If not, we set that so that before 1.0, features increment minor not major.

With this setup, every change merged triggers either no version change (if no feat/fix) or a properly bumped new version with automated release of image and chart. This satisfies the requirement *“Release Please for versioning the docker image and for the helm releases”* – we will have **automated semantic versioning** across the board: - **Docker image tags** versioned by Release Please (with the version coming from package.json and commits). - **Helm chart version** bumped in Chart.yaml by Release Please[[31]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=Helm%20Charts). - Both artifacts released/published for each version.

## Conclusion and Next Steps

We’ve outlined a comprehensive plan for setting up the development workspaces and the CI/CD automation:

* **Dockerfile**: Use Debian, install Git, GH CLI[[3]](https://www.boxpiper.com/posts/github-cli#:~:text=,hostname%20%3Chostname), Node 20, Claude CLI, Happy CLI. Ensure entrypoint script is in place for cloning logic and starting happy/claude.
* **Helm Chart**: Add values for workspace.repos (list of repo URLs), workspace.storageSize (size for code volume), and githubAuth.pvcSize (size for GH config PVC). Provide defaults (e.g., repos default to happy-cli, 10Gi storage, 1Gi PVC). Template an initContainer (for cloning) and the necessary volumes (emptyDir and PVC) into the Deployment spec. Set the container’s workingDir to /workspace. Mount volumes at /workspace and /root/.config/gh. Also allow overriding image tag if needed (though automated by releases anyway).
* **InitContainer Script**: Possibly generate it from the values (Helm can embed a script with the repo list). Or use a small helper container image that can interpret a list. One idea: pass the list via environment or as an encoded JSON, and have the init container run a loop. Simpler: use wget/git in a shell loop. (We can write a one-liner shell in Helm, but careful with quotes and such.)
* **Security**: If needed, set the init container’s command to run as root to ensure it can write to the volume (or set fsGroup on the volume).
* **GitHub Auth**: Create K8s Secret for PAT, mount as env, configure GH login in entrypoint or init. Persist config on PVC.
* **Claude Auth**: Document that an initial login is required unless we find an API-key method. We might simply log a message on startup: “Please run claude /login and follow the steps to authenticate Claude Code.” If a user or admin attaches to the pod (via Happy CLI’s QR from mobile or via kubectl exec) they could do that once. If persistent, it might carry over. (This part might require a more elegant solution in future.)
* **Release Automation**: Add release-please config file. Mark version lines with comments[[28]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=,version). Set up GitHub Actions:
* one for release-please (to open PRs),
* one for building/pushing Docker on tag,
* one for packaging/pushing Helm chart on tag.
* **Example Documentation**: We should update the README or docs to instruct how to configure values (e.g., how to specify multiple repos, how to supply the bot token secret, etc.). Possibly, provide an example values.yaml snippet:
* workspace:
   repos:
   - "https://github.com/slopus/happy-cli.git"
   - "https://github.com/example/other-repo.git"
   storageSize: 10Gi
  githubAuth:
   pvcSize: 1Gi
* and explain that by default happy-cli is cloned if no repos given.
* **Testing**: We will test the helm chart by deploying a workspace to ensure:
* The init container clones repos correctly (check /workspace contents).
* The GH auth persists (e.g., do gh auth status inside pod to confirm).
* The happy CLI can start Claude (it will prompt for Claude login likely).
* Making a trivial commit in the workspace and pushing works (to validate git + gh).
* The Release Please action triggers on commit messages. We might simulate conventional commits and see that a PR is opened bumping versions in package.json and Chart.yaml.

By providing all these details and linking to relevant documentation and examples, a future developer or AI agent should have a **solid roadmap** to implement the feature. They can refer to the embedded quotes for guidance on specific points (e.g., how to mount volumes[[15]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=Persistent%20volumes%20allow%20you%20to,hard%20drive%20on%20every%20restart)[[24]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=emptyDir%3A%20%20%20%20,Kubernetes%20EmptyDirVolumeSource%20v1), how release-please bumps chart versions[[31]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=Helm%20Charts), how conventional commits map to semver[[25]](https://github.com/googleapis/release-please#:~:text=The%20most%20important%20prefixes%20you,should%20have%20in%20mind%20are), and how to perform headless GH login[[12]](https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode#:~:text=The%20solution%20is%20simple%3A)).

In summary, this design will enable spinning up robust, multi-repo coding workspaces with all necessary tools (Claude Code and Happy CLI) and a fully automated release pipeline for versioning the resulting software artifacts. The next steps are to implement the Dockerfile and Helm chart changes accordingly, and set up the GitHub Actions workflows as described. With these in place, any new code or changes can be managed and released efficiently, and the AI agent will have a rich environment to operate in.

[[1]](https://claude.com/blog/workspaces#:~:text=We%27re%20introducing%20Workspaces%20in%20the,on%20a%20more%20granular%20level) [[2]](https://claude.com/blog/workspaces#:~:text=For%20developers%20using%20Claude%20across,organization%20and%20individual%20API%20keys) Workspaces in the Anthropic API Console | Claude

<https://claude.com/blog/workspaces>

[[3]](https://www.boxpiper.com/posts/github-cli#:~:text=,hostname%20%3Chostname) [[11]](https://www.boxpiper.com/posts/github-cli#:~:text=,setup) GitHub CLI - GitHub and command line in 2025 - Box Piper

<https://www.boxpiper.com/posts/github-cli>

[[4]](https://github.com/slopus/happy-cli#:~:text=Requirements) [[8]](https://github.com/slopus/happy-cli#:~:text=Installation) [[9]](https://github.com/slopus/happy-cli#:~:text=device%20github,cli) [[10]](https://github.com/slopus/happy-cli#:~:text=This%20will%3A) [[18]](https://github.com/slopus/happy-cli#:~:text=Commands) [[19]](https://github.com/slopus/happy-cli#:~:text=,command%20available%20in%20PATH) GitHub - slopus/happy-cli: Happy Coder CLI to connect your local Claude Code to mobile device

<https://github.com/slopus/happy-cli>

[[5]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=To%20install%20Claude%20Code%2C%20log,and%20run%20the%20following%20command) [[6]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=curl%20) [[17]](https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/#:~:text=Logging%20in) Installing Claude Code on a VPS at Hostinger - Hostinger Help Center

<https://www.hostinger.com/support/11929523-installing-claude-code-on-a-vps-at-hostinger/>

[[7]](https://shipyard.build/blog/claude-code-on-the-web/#:~:text=CC%20web%20also%20has%20a,when%20you%20might%20be%20AFK) Shipyard | Claude Code on the web: get cloud-hosted, secure agent runners

<https://shipyard.build/blog/claude-code-on-the-web/>

[[12]](https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode#:~:text=The%20solution%20is%20simple%3A) [[13]](https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode#:~:text=1.%20Run%20,time%20code%204.%20Done) github cli - How to do `gh auth login` when run in headless mode? - Stack Overflow

<https://stackoverflow.com/questions/78890002/how-to-do-gh-auth-login-when-run-in-headless-mode>

[[14]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=Generally%2C%20containers%20are%20stateless,removed%20when%20the%20container%20restarts) [[15]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=Persistent%20volumes%20allow%20you%20to,hard%20drive%20on%20every%20restart) [[24]](https://www.devspace.sh/component-chart/docs/guides/persistent-volumes#:~:text=emptyDir%3A%20%20%20%20,Kubernetes%20EmptyDirVolumeSource%20v1) Persistent Volumes | Component Helm Chart | Documentation

<https://www.devspace.sh/component-chart/docs/guides/persistent-volumes>

[[16]](https://developer.1password.com/docs/cli/shell-plugins/github/#:~:text=Use%201Password%20to%20authenticate%20the,config%2Fgh%2Fhosts.yml.%20Next%20steps) Use 1Password to authenticate the GitHub CLI with biometrics

<https://developer.1password.com/docs/cli/shell-plugins/github/>

[[20]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=initContainers%3A%20,%27https%3A%2F%2F%24%28GIT_USERNAME%29%3A%24%28GIT_PASSWORD%29%40gitlab.company.com%3E%2Fpath%2Fto%2Frepo.git) [[21]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=env%3A%20,GIT_PASSWORD%20valueFrom%3A%20secretKeyRef%3A%20key%3A%20password) [[22]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=,%27https%3A%2F%2F%24%28GIT_USERNAME%29%3A%24%28GIT_PASSWORD%29%40gitlab.company.com%3E%2Fpath%2Fto%2Frepo.git) [[23]](https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238#:~:text=valueFrom%3A%20secretKeyRef%3A%20key%3A%20username%20name%3A,secret) Cloning git repos using Kubernetes initContainers and Secrets | by Stefvnf | Medium

<https://stefvnf.medium.com/cloning-git-repos-using-kubernetes-initcontainers-and-secrets-8609e3b2d238>

[[25]](https://github.com/googleapis/release-please#:~:text=The%20most%20important%20prefixes%20you,should%20have%20in%20mind%20are) [[26]](https://github.com/googleapis/release-please#:~:text=,result%20in%20a%20SemVer%20major) [[27]](https://github.com/googleapis/release-please#:~:text=1,Release%20based%20on%20the%20tag) [[34]](https://github.com/googleapis/release-please#:~:text=When%20the%20Release%20PR%20is,please%20takes%20the%20following%20steps) [[35]](https://github.com/googleapis/release-please#:~:text=,a%20convention%20for%20publication%20tooling) GitHub - googleapis/release-please: generate release PRs based on the conventionalcommits.org spec

<https://github.com/googleapis/release-please>

[[28]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=,version) [[29]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=%7B%20,false) [[30]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=The%20file%20gets%20its%20version,Add%20the%20annotation) [[39]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=The%20configuration%20path%20must%20point,a%20directory%20containing%20versionable%20files) [[40]](https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/#:~:text=annotation%3A) What Release-Please Can't Do - Adaptive Enforcement Lab

<https://adaptive-enforcement-lab.com/blog/2025/11/30/what-release-please-cant-do/>

[[31]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=Helm%20Charts) [[32]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=3.%20Release,the%20packaged%20helm%20charts%20are) [[33]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=4,the%20packaged%20helm%20charts%20are) [[36]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=GitHub%20container%20registry%2C%20Lerna%20will,0.4.2) [[37]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=tag%20all%20images%3A%20,be%20overwritten%20on%20every%20release) [[38]](https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md#:~:text=6,router) cosmo/docs/releasing.md at main · wundergraph/cosmo · GitHub

<https://github.com/wundergraph/cosmo/blob/main/docs/releasing.md>
