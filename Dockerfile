FROM debian:bookworm-slim

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 24
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN (curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg) \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN curl -fsSL https://claude.ai/install.sh | bash

# Enable Corepack and set up pnpm environment
RUN corepack enable

# Set pnpm environment variables
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT="0"

# Prepare pnpm using Corepack and install Happy CLI
RUN corepack prepare pnpm@10.26.0 --activate && \
    mkdir -p $PNPM_HOME && \
    pnpm install -g happy-coder

# Create workspace directory
RUN mkdir -p /workspace /scripts

# Copy entrypoint and helper scripts
COPY scripts/entrypoint.sh /scripts/entrypoint.sh
COPY scripts/clone-repos.sh /usr/local/bin/clone-repos

RUN chmod +x /scripts/entrypoint.sh /usr/local/bin/clone-repos

# Create non-root user with UID 1001 (following Bitnami pattern)
# Using 1001 instead of 1000 to avoid conflicts with host users
RUN groupadd --gid 1001 coder && \
    useradd --uid 1001 --gid 1001 --shell /bin/bash --create-home coder

# Create directories that the user needs write access to
RUN mkdir -p /home/coder/.config/gh /home/coder/.claude && \
    chown -R 1001:1001 /home/coder && \
    chown -R 1001:1001 /workspace

# Switch to non-root user
USER 1001

# Set working directory
WORKDIR /workspace

# Default entrypoint
ENTRYPOINT ["/scripts/entrypoint.sh"]
# Run happy with pseudo-TTY via script command
# -q: quiet (no "Script started" messages)
# -e: return exit code of child process
# -f: flush output immediately (real-time logs)
# -c: run command instead of shell
CMD ["script", "-qefc", "happy --no-qr", "/dev/null"]
